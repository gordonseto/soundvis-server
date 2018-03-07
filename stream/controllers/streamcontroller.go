package stream

import (
	"gopkg.in/mgo.v2"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/stream/IO"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stations/controllers"
	"encoding/json"
	"github.com/gordonseto/soundvis-server/users/controllers"
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

	response.CurrentStation, err = getCurrentStation(user.CurrentPlaying, sc.session)
	if err != nil {
		panic(err)
	}

	response.CurrentStreamURL = getCurrentStreamURL(user.CurrentPlaying, response.CurrentStation)

	basecontroller.SendResponse(w, response)
}

func (sc *StreamController) SetCurrentStream(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// check if authenticated
	user, err := authentication.CheckAuthentication(r, sc.session)
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
	station, err := getCurrentStation(request.CurrentStream, sc.session)
	if err != nil {
		panic(err)
	}

	// set user's values to match request
	user.IsPlaying = request.IsPlaying
	user.CurrentPlaying = request.CurrentStream

	// save user into db
	err = users.UpdateUser(sc.session, user)
	if err != nil {
		panic(err)
	}

	// create response
	response := streamIO.GetCurrentStreamResponse{}
	response.IsPlaying = user.IsPlaying
	response.CurrentStation = station
	response.CurrentStreamURL = getCurrentStreamURL(user.CurrentPlaying, station)

	basecontroller.SendResponse(w, response)
}

// takes in currentPlaying and returns the station corresponding with that id
// currentPlaying is an id to a station or recording
func getCurrentStation(currentPlaying string, session *mgo.Session) (*models.Station, error) {
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

func getCurrentStreamURL(currentPlaying string, currentStation *models.Station) string {
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
	return false
}