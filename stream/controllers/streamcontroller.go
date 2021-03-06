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
	"strconv"
	"net/http/httputil"
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

func (sc *StreamController) DE1PostPath() string {
	return "/stream/:isPlaying/:currentStream/:currentVolume"
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
	// parse request
	request := streamIO.SetCurrentStreamRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	sc.handleSetCurrentStreamRequest(w, r, &request)
}

// this is not good practice to put an update on a GET, but this is the easiest way for DE1 to update stream
func (sc *StreamController) SetCurrentStreamDE1(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))

	request := streamIO.SetCurrentStreamRequest{}
	if p.ByName("isPlaying") == "1" {
		request.IsPlaying = true
	}
	request.CurrentStream = p.ByName("currentStream")
	volume, err := strconv.Atoi(p.ByName("currentVolume"))
	if err != nil {
		volume = 0
	}
	request.CurrentVolume = volume

	// TODO: Remove this
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}
	request.CurrentStream = user.CurrentPlaying

	sc.handleSetCurrentStreamRequest(w, r, &request)
}

func (sc *StreamController) handleSetCurrentStreamRequest(w http.ResponseWriter, r *http.Request, request *streamIO.SetCurrentStreamRequest) {
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

	response, err := streamcontrollerhelper.UpdateUsersStream(request, user)
	if err != nil {
		panic(err)
	}

	// if request was from DE1, send notification to Android device
	if userAgent == authentication.DE1 {
		err = notifications.SendStreamUpdateNotification([]string{user.DeviceToken}, *response)
		fmt.Println(err)
	} else if userAgent == authentication.ANDROID {
		// else send socket message to raspberry pi
		err = socketmanager.Shared().SendStreamUpdateMessage(user.Id.Hex(), *response)
		if err != nil {
			fmt.Println("Socket error for userId: " + user.Id.Hex())
			fmt.Println(err)
		}
		// send socket message to DE1
		err = socketmanager.Shared().SendStreamUpdateMessage(user.Id.Hex() + "DE1", *response)
		if err != nil {
			fmt.Println("Socket error for userId: " + user.Id.Hex() + "DE1")
			fmt.Println(err)
		}
	}

	basecontroller.SendResponse(w, response)
}