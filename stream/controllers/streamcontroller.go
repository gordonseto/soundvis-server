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
)

type (
	StreamController struct {
		session *mgo.Session
	}
)

func (sc StreamController) GETPath() string {
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

	user.CurrentPlaying = "57599"
	response.CurrentStation, err = getCurrentStation(user.CurrentPlaying, sc.session)
	if err != nil {
		panic(err)
	}

	basecontroller.SendResponse(w, response)
}

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

func currentPlayingIsRecording(currentPlaying string) bool {
	return false
}