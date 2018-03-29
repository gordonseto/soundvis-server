package stream

import (
	"encoding/json"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/notifications"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/gordonseto/soundvis-server/users/repositories"
	"fmt"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"github.com/gordonseto/soundvis-server/socketmanager"
	"time"
	"github.com/gordonseto/soundvis-server/listeningsessions/repositories"
	"log"
)

type (
	StreamController struct {
	}
)

func (sc *StreamController) GETPath() string {
	return "/stream"
}

func (sc *StreamController) POSTPath() string {
	return "/stream"
}

func NewStreamController() *StreamController {
	return &StreamController{}
}

func (sc *StreamController) GetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying
	response.CurrentPlaying = user.CurrentPlaying

	response.CurrentStation, response.CurrentSong, err =  streamhelper.GetCurrentStationAndSongPlaying(user.CurrentPlaying, time.Now().Unix() - user.StreamUpdatedAt)
	if err != nil {
		panic(err)
	}

	response.CurrentStreamURL = streamhelper.GetStreamURL(user.CurrentPlaying, response.CurrentStation)
	err = streamhelper.GetImageURLForSong(response.CurrentSong)
	if err != nil {
		log.Println(err)
	}

	basecontroller.SendResponse(w, response)
}

func (sc *StreamController) SetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// check if authenticated
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	// check if valid user-agent
	userAgent, err := authentication.GetUserAgent(r)
	if err != nil {
		panic(err)
	}

	// parse request
	request := streamIO.SetCurrentStreamRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// check if stream is valid
	station, err := streamhelper.GetStation(request.CurrentStream)
	if err != nil {
		panic(err)
	}

	// hold onto the user's previous currentPlaying and streamUpdatedAt to save as a listening session
	previousIsPlaying := user.IsPlaying
	previousPlaying := user.CurrentPlaying
	previousStreamUpdatedAt := user.StreamUpdatedAt

	// set user's values to match request
	user.IsPlaying = request.IsPlaying
	user.CurrentPlaying = request.CurrentStream

	// set the user's streamUpdatedAt
	user.StreamUpdatedAt = time.Now().Unix()

	// create response
	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying
	response.CurrentPlaying = user.CurrentPlaying
	response.CurrentStation = station
	response.CurrentStreamURL = streamhelper.GetStreamURL(user.CurrentPlaying, station)
	response.CurrentSong, err = streamhelper.GetCurrentSongPlaying(user.CurrentPlaying, time.Now().Unix() - user.StreamUpdatedAt, station)
	if err != nil {
		panic(err)
	}
	err = streamhelper.GetImageURLForSong(response.CurrentSong)
	if err != nil {
		log.Println(err)
	}

	// update user in db
	err = usersrepository.Shared().UpdateUser(user)
	if err != nil {
		panic(err)
	}

	// save previous listening session if needed
	if previousIsPlaying {
		err = listeningsessionsrepository.Shared().InsertListeningSession(user.Id.Hex(), previousPlaying, time.Now().Unix() - previousStreamUpdatedAt, previousStreamUpdatedAt)
		if err != nil {
			log.Println("Error saving user's listening session: ")
			log.Println(err)
		}
	}

	// if request was from DE1, send notification to Android device
	if userAgent == authentication.DE1 {
		err = notifications.SendStreamUpdateNotification([]string{user.DeviceToken}, response)
		fmt.Println(err)
	} else if userAgent == authentication.ANDROID {
		// else send socket message to DE1 connection
		err = socketmanager.Shared().SendStreamUpdateMessage(user.Id.Hex(), response)
		if err != nil {
			fmt.Println("Socket error for userId: " + user.Id.Hex())
			fmt.Println(err)
		}
	}

	basecontroller.SendResponse(w, response)
}