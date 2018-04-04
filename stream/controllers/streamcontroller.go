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
	r.ParseForm()
	for key, value := range r.Form {
		fmt.Println("%s = %s\n", key, value)
	}

	log.Println("New request at /stream:")
	log.Println(r)

	log.Println("BODY:")
	log.Println(r.Body)
	log.Println("End body")
	// check if authenticated
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	log.Println("Found user:")
	log.Println(user.Id)
	// check if valid user-agent
	userAgent, err := authentication.GetUserAgent(r)
	if err != nil {
		panic(err)
	}


	log.Println("User agent:")
	log.Println(userAgent)

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))

	// parse request
	request := streamIO.SetCurrentStreamRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}
	log.Println("Request parsed:")
	log.Println(request)

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

func (sc *StreamController) SetCurrentStreamDE1(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Println(p.ByName("isPlaying"))
	log.Println(p.ByName("currentStream"))
	log.Println(p.ByName("currentVolume"))

	log.Println("Hello world!")
}