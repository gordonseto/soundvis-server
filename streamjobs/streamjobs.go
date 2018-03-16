package streamjobs

import (
	models2 "github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"github.com/gordonseto/soundvis-server/stream/models"
	"github.com/gordonseto/soundvis-server/streammanager"
	"github.com/gordonseto/soundvis-server/users/models"
	"github.com/gordonseto/soundvis-server/users/repositories"
	"log"
	"github.com/gordonseto/soundvis-server/notifications"
	"github.com/gordonseto/soundvis-server/socketmanager"
)

type StreamJobManager struct {
	userRepository     *usersrepository.UsersRepository
	streamManager      *streammanager.StreamManager
	socketManager *socketmanager.SocketManager
	previousPlayingMap map[string]string
}

func NewStreamJobManager(ur *usersrepository.UsersRepository, stm *streammanager.StreamManager, skm *socketmanager.SocketManager) *StreamJobManager {
	return &StreamJobManager{ur, stm, skm, make(map[string]string)}
}

func StreamJobName() string {
	return "stream_job"
}

// fetches currentPlaying for all users with currentPlaying = true
// sends socket message and android notification if song has changed
func (sjm *StreamJobManager) RefreshNowPlaying() {
	users, err := sjm.userRepository.FindUsersByIsPlaying(true)
	if err != nil {
		log.Println(err)
		return
	}
	for _, user := range users {
		go sjm.CheckNowPlayingForUser(user)
	}
}

// checks user's currentPlaying; if the station or song has changed, sends a socket message
// and notification to user
func (sjm *StreamJobManager) CheckNowPlayingForUser(user models.User) {
	// get the current station and song playing for user
	station, song, err := sjm.streamManager.GetCurrentStationAndSongPlaying(user.CurrentPlaying)
	if err != nil {
		log.Println(err)
		return
	}

	// concatenate into string
	stringified := stationAndSongToString(station, song)
	// get station and song already stored in map for user
	previousPlaying, _ := sjm.previousPlayingMap[user.Id.Hex()]

	// if previousPlaying has changed, a new song is playing
	if previousPlaying != stringified {
		response := streamIO.GetCurrentStreamResponse{}
		response.IsPlaying = user.IsPlaying
		response.CurrentStation = station
		response.CurrentSong = song
		response.CurrentStreamURL = sjm.streamManager.GetStreamURL(user.CurrentPlaying, station)
		log.Println(response.CurrentStation.Name + " - " + response.CurrentSong.Name)

		// send android notification
		notifications.SendStreamUpdateNotification([]string{user.DeviceToken}, response)
		// send socket message
		sjm.socketManager.SendStreamUpdateMessage(user.Id.Hex(), response)
	}
	// set previousPlaying to current song
	sjm.previousPlayingMap[user.Id.Hex()] = stringified
}

// concatenates together station and song for comparison
func stationAndSongToString(station *models2.Station, song *stream.Song) string {
	return station.Id.Hex() + song.Name + song.Title
}
