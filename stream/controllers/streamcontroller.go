package stream

import (
	"encoding/json"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/config"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/notifications"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"github.com/gordonseto/soundvis-server/stream/models"
	"github.com/gordonseto/soundvis-server/users/controllers"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"net/http"
	"strings"
	"fmt"
	"encoding/xml"
	"github.com/gordonseto/soundvis-server/stations/controllers"
)

type (
	StreamController struct {
		session *mgo.Session
	}
)

func (sc *StreamController) GETPath() string {
	return "/stream"
}

func (sc *StreamController) POSTPath() string {
	return "/stream"
}

func NewStreamController(s *mgo.Session) *StreamController {
	return &StreamController{s}
}

func (sc *StreamController) GetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r, sc.session)
	if err != nil {
		panic(err)
	}

	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying

	response.CurrentStation, err = getStation(user.CurrentPlaying, sc.session)
	if err != nil {
		panic(err)
	}

	response.CurrentStreamURL = getStreamURL(user.CurrentPlaying, response.CurrentStation)

	response.CurrentSong, err = getCurrentSongPlaying(user.CurrentPlaying)
	if err != nil {
		panic(err)
	}

	basecontroller.SendResponse(w, response)
}

func (sc *StreamController) SetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// check if authenticated
	user, err := authentication.CheckAuthentication(r, sc.session)
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
	station, err := getStation(request.CurrentStream, sc.session)
	if err != nil {
		panic(err)
	}

	// set user's values to match request
	user.IsPlaying = request.IsPlaying
	user.CurrentPlaying = request.CurrentStream

	// update user in db
	err = users.UpdateUser(sc.session, user)
	if err != nil {
		panic(err)
	}

	// create response
	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying
	response.CurrentStation = station
	response.CurrentStreamURL = getStreamURL(user.CurrentPlaying, station)
	response.CurrentSong, err = getCurrentSongPlaying(user.CurrentPlaying)
	if err != nil {
		panic(err)
	}

	// if request was from DE1, send notification to Android device
	if userAgent == authentication.DE1 {
		notifications.SendStreamUpdateNotification([]string{user.DeviceToken}, response)
	} else if userAgent == authentication.ANDROID {
		// TODO: Implement this
	}

	basecontroller.SendResponse(w, response)
}

// takes in currentPlaying and returns the station corresponding with that id
// currentPlaying is an id to a station or recording
func getStation(currentPlaying string, session *mgo.Session) (*models.Station, error) {
	if currentPlaying == "" {
		return nil, nil
	}
	if currentPlayingIsRecording(currentPlaying) {
		// TODO: Implement this
		return nil, nil
	} else {
		stationsController := stations.NewStationsController(session)
		stations, err := stations.GetStations(stationsController, currentPlaying, 0, 0)
		if err != nil || len(stations) < 1 {
			return nil, err
		}
		return &stations[0], nil
	}
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

// gets the most recent playing song in currentPlaying
// currentPlaying is the id of a station or recording
func getCurrentSongPlaying(currentPlaying string) (*stream.Song, error) {
	if currentPlaying == "" {
		return nil, nil
	}

	if currentPlayingIsRecording(currentPlaying) {
		// TODO: Implement this
		return nil, nil
	} else {
		// build url
		url := "http://api.dirble.com/v2/station/" + currentPlaying + "/song_history?token=" + config.DIRBLE_API_KEY
		songs := make([]stream.Song, 0)

		// make request
		body, err := basecontroller.MakeRequest(url, http.MethodGet, 10)
		if err != nil {
			return nil, err
		}

		// parse request
		err = json.Unmarshal(body, &songs)
		if err != nil {
			return nil, err
		}

		if len(songs) < 1 {
			return nil, nil
		}

		return &songs[0], nil
	}
}

type html struct {
	Body body `xml:"body"`
}
type body struct {
	Content string `xml:",innerxml"`
}

// takes in a shoutcast streamURL and returns the current song playing if valid
func GetCurrentSongPlayingShoutcast(streamURL string) (string, error) {
	// remove trailing url component
	streamBaseArray := strings.Split(streamURL, "/")
	streamBase := ""
	for j := 0; j < len(streamBaseArray)-1; j++ {
		streamBase += streamBaseArray[j] + "/"
	}
	fmt.Println(streamBase)
	// send request to urlbase + "7"
	// if this request succeeds, use this station, else discard it
	res, err := basecontroller.MakeRequest(streamBase + "7", http.MethodGet, 5)
	if err != nil {
		return "", err
	} else {
		fmt.Println(res)
		h := html{}
		err := xml.NewDecoder(strings.NewReader(string(res))).Decode(&h)
		if err != nil {
			return "", err
		} else {
			return h.Body.Content, nil
		}
	}
}

func currentPlayingIsRecording(currentPlaying string) bool {
	if currentPlaying == "" {
		return false
	}
	return false
}
