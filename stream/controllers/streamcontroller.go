package stream

import (
	"encoding/json"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/notifications"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"fmt"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"github.com/gordonseto/soundvis-server/socketmanager"
	"time"
	"log"
	"github.com/gordonseto/soundvis-server/stream/helpers"
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
	response.CurrentVolume = user.CurrentVolume

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

	response, err := streamcontrollerhelper.UpdateUsersStream(&request, user)
	if err != nil {
		panic(err)
	}

	// if request was from DE1, send notification to Android device
	if userAgent == authentication.DE1 {
		err = notifications.SendStreamUpdateNotification([]string{user.DeviceToken}, *response)
		fmt.Println(err)
	} else if userAgent == authentication.ANDROID {
		// else send socket message to DE1 connection
		err = socketmanager.Shared().SendStreamUpdateMessage(user.Id.Hex(), *response)
		if err != nil {
			fmt.Println("Socket error for userId: " + user.Id.Hex())
			fmt.Println(err)
		}
	}

	basecontroller.SendResponse(w, response)
}