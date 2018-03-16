package streamjobmanager

import (
	models2 "github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"github.com/gordonseto/soundvis-server/stream/models"
	"github.com/gordonseto/soundvis-server/users/models"
	"log"
	"github.com/gordonseto/soundvis-server/notifications"
	"sync"
	"github.com/gordonseto/soundvis-server/users/repositories"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"github.com/gordonseto/soundvis-server/socketmanager"
	"fmt"
)

type StreamJobManager struct {
	previousPlayingMap map[string]string
}

var instance *StreamJobManager
var once sync.Once
var mutex = &sync.Mutex{}

func Shared() *StreamJobManager {
	once.Do(func() {
		instance = &StreamJobManager{make(map[string]string)}
	})
	return instance
}

func StreamJobName() string {
	return "stream_job"
}

// fetches currentPlaying for all users with currentPlaying = true
// sends socket message and android notification if song has changed
func (sjm *StreamJobManager) RefreshNowPlaying() {
	users, err := usersrepository.Shared().FindUsersByIsPlaying(true)
	if err != nil {
		log.Println(err)
		return
	}
	for _, user := range users {
		fmt.Println(user.Id.Hex())
		go sjm.checkNowPlayingForUser(user)
	}
}

// checks user's currentPlaying; if the station or song has changed, sends a socket message
// and notification to user
func (sjm *StreamJobManager) checkNowPlayingForUser(user models.User) {
	// get the current station and song playing for user
	station, song, err := streamhelper.GetCurrentStationAndSongPlaying(user.CurrentPlaying)
	if err != nil {
		log.Println(err)
		return
	}

	// concatenate into string
	stringified := stationAndSongToString(station, song)
	// get station and song already stored in map for user
	mutex.Lock()
	previousPlaying, _ := sjm.previousPlayingMap[user.Id.Hex()]
	// set previousPlaying to current song
	sjm.previousPlayingMap[user.Id.Hex()] = stringified
	mutex.Unlock()

	// if previousPlaying has changed, a new song is playing
	if previousPlaying != stringified {
		response := streamIO.GetCurrentStreamResponse{}
		response.IsPlaying = user.IsPlaying
		response.CurrentStation = station
		response.CurrentSong = song
		response.CurrentStreamURL = streamhelper.GetStreamURL(user.CurrentPlaying, station)
		log.Println(response.CurrentStation.Name + ", " + response.CurrentSong.Name + " - " + response.CurrentSong.Title)

		// send android notification
		notifications.SendStreamUpdateNotification([]string{user.DeviceToken}, response)
		// send socket message
		socketmanager.Shared().SendStreamUpdateMessage(user.Id.Hex(), response)
	}
}

// concatenates together station and song for comparison
func stationAndSongToString(station *models2.Station, song *stream.Song) string {
	return station.Id.Hex() + song.Name + song.Title
}
