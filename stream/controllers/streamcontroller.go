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
	"github.com/gordonseto/soundvis-server/socketmanager"
	"github.com/gordonseto/soundvis-server/streammanager"
)

type (
	StreamController struct {
		usersRepository *usersrepository.UsersRepository
		socketManager *socketmanager.SocketManager
		streamManager *streammanager.StreamManager
	}
)

func (sc *StreamController) GETPath() string {
	return "/stream"
}

func (sc *StreamController) POSTPath() string {
	return "/stream"
}

func NewStreamController(ur *usersrepository.UsersRepository, skm *socketmanager.SocketManager, stm *streammanager.StreamManager) *StreamController {
	return &StreamController{ur, skm, stm}
}

func (sc *StreamController) GetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r, sc.usersRepository)
	if err != nil {
		panic(err)
	}

	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying

	response.CurrentStation, response.CurrentSong, err =  sc.streamManager.GetCurrentStationAndSongPlaying(user.CurrentPlaying)
	if err != nil {
		panic(err)
	}

	response.CurrentStreamURL = sc.streamManager.GetStreamURL(user.CurrentPlaying, response.CurrentStation)

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
	station, err := sc.streamManager.GetStation(request.CurrentStream)
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
	response.CurrentStreamURL = sc.streamManager.GetStreamURL(user.CurrentPlaying, station)
	response.CurrentSong, err = sc.streamManager.GetCurrentSongPlaying(user.CurrentPlaying, station)
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