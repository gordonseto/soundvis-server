package stream

import (
	"encoding/json"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/notifications"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/gordonseto/soundvis-server/users/repositories"
	"github.com/gordonseto/soundvis-server/stations/repositories"
	"github.com/gordonseto/soundvis-server/stream/models"
	"github.com/gordonseto/soundvis-server/stationsfetcher"
	"fmt"
	"github.com/gordonseto/soundvis-server/socketmanager"
)

type (
	StreamController struct {
		usersRepository *usersrepository.UsersRepository
		stationsRepository *stationsrepository.StationsRepository
		socketManager *socketmanager.SocketManager
	}
)

func (sc *StreamController) GETPath() string {
	return "/stream"
}

func (sc *StreamController) POSTPath() string {
	return "/stream"
}

func NewStreamController(ur *usersrepository.UsersRepository, sr *stationsrepository.StationsRepository, sm *socketmanager.SocketManager) *StreamController {
	return &StreamController{ur, sr, sm}
}

func (sc *StreamController) GetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r, sc.usersRepository)
	if err != nil {
		panic(err)
	}

	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying

	response.CurrentStation, err =  sc.getStation(user.CurrentPlaying)
	if err != nil {
		panic(err)
	}

	response.CurrentSong, err = sc.getCurrentSongPlaying(user.CurrentPlaying, response.CurrentStation)
	if err != nil {
		panic(err)
	}

	response.CurrentStreamURL = getStreamURL(user.CurrentPlaying, response.CurrentStation)

	basecontroller.SendResponse(w, response)
}

func (sc *StreamController) SetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// check if authenticated
	user, err := authentication.CheckAuthentication(r, sc.usersRepository)
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
	station, err := sc.getStation(request.CurrentStream)
	if err != nil {
		panic(err)
	}

	// set user's values to match request
	user.IsPlaying = request.IsPlaying
	user.CurrentPlaying = request.CurrentStream

	// update user in db
	err = sc.usersRepository.UpdateUser(user)
	if err != nil {
		panic(err)
	}

	// create response
	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying
	response.CurrentStation = station
	response.CurrentStreamURL = getStreamURL(user.CurrentPlaying, station)
	response.CurrentSong, err = sc.getCurrentSongPlaying(user.CurrentPlaying, station)
	if err != nil {
		panic(err)
	}

	// if request was from DE1, send notification to Android device
	if userAgent == authentication.DE1 {
		err = notifications.SendStreamUpdateNotification([]string{user.DeviceToken}, response)
		fmt.Println(err)
	} else if userAgent == authentication.ANDROID {
		// else send socket message to DE1 connection
		userId := "3"
		err = sc.socketManager.SendStreamUpdateMessage(userId, response)
		if err != nil {
			fmt.Println("Socket error for userId: " + userId)
			fmt.Println(err)
		}
	}

	basecontroller.SendResponse(w, response)
}

// gets the audio stream url from currentPlaying
// currentPlaying is the id of a station or recording
func getStreamURL(currentPlaying string, currentStation *models.Station) string {
	if currentPlaying == "" {
		return ""
	}
	if currentPlayingIsRecording(currentPlaying) {
		// TODO: Implement this
		return ""
	} else {
		return currentStation.StreamURL
	}
}

func currentPlayingIsRecording(currentPlaying string) bool {
	if currentPlaying == "" {
		return false
	}
	return false
}

// currentPlaying is an id for a station or recording
func (sc *StreamController) getStation(currentPlaying string) (*models.Station, error) {
	if currentPlaying == "" {
		return nil, nil
	}

	if currentPlayingIsRecording(currentPlaying) {
		// TODO: Implement this
		return nil, nil
	} else {
		return sc.stationsRepository.FindStationById(currentPlaying)
	}
}

// currentPlaying is an id for a station or recording
func (sc *StreamController) getCurrentSongPlaying(currentPlaying string, station *models.Station) (*stream.Song, error) {
	if currentPlaying == "" {
		return nil, nil
	}

	if currentPlayingIsRecording(currentPlaying) {
		// TODO: Implement this
		return nil, nil
	} else {
		return stationsfetcher.GetCurrentSongPlayingShoutcast(station.StreamURL)
	}
}