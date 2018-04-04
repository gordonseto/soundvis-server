package streamcontrollerhelper

import (
	"github.com/gordonseto/soundvis-server/stream/IO"
	"time"
	"log"
	"github.com/gordonseto/soundvis-server/users/repositories"
	"github.com/gordonseto/soundvis-server/listeningsessions/repositories"
	"github.com/gordonseto/soundvis-server/users/models"
	"github.com/gordonseto/soundvis-server/streamhelper"
)

func UpdateUsersStream(request *streamIO.SetCurrentStreamRequest, user *models.User) (*streamIO.GetCurrentStreamResponse, error) {
	// check if stream is valid
	station, err := streamhelper.GetStation(request.CurrentStream)
	if err != nil {
		return nil, err
	}

	// hold onto the user's previous currentPlaying and streamUpdatedAt to save as a listening session
	previousIsPlaying := user.IsPlaying
	previousPlaying := user.CurrentPlaying
	previousStreamUpdatedAt := user.StreamUpdatedAt

	// set user's values to match request
	user.IsPlaying = request.IsPlaying
	user.CurrentPlaying = request.CurrentStream
	user.CurrentVolume = request.CurrentVolume

	// set the user's streamUpdatedAt if isPlaying or CurrentPlaying has changed
	// if only volume has changed, do not update streamUpdatedAt
	if previousIsPlaying != user.IsPlaying || previousPlaying != user.CurrentPlaying {
		user.StreamUpdatedAt = time.Now().Unix()
	}

	// create response
	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying
	response.CurrentPlaying = user.CurrentPlaying
	response.CurrentVolume = user.CurrentVolume
	response.CurrentStation = station
	response.CurrentStreamURL = streamhelper.GetStreamURL(user.CurrentPlaying, station)
	response.CurrentSong, err = streamhelper.GetCurrentSongPlaying(user.CurrentPlaying, time.Now().Unix() - user.StreamUpdatedAt, station)
	if err != nil {
		return nil, err
	}
	err = streamhelper.GetImageURLForSong(response.CurrentSong)
	if err != nil {
		log.Println(err)
	}

	// update user in db
	err = usersrepository.Shared().UpdateUser(user)
	if err != nil {
		return nil, err
	}

	// save previous listening session if needed
	if previousIsPlaying {
		err = listeningsessionsrepository.Shared().InsertListeningSession(user.Id.Hex(), previousPlaying, time.Now().Unix() - previousStreamUpdatedAt, previousStreamUpdatedAt)
		if err != nil {
			log.Println("Error saving user's listening session: ")
			log.Println(err)
		}
	}

	return &response, nil
}