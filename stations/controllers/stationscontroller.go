package stations

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/stations/IO"
	"github.com/gordonseto/soundvis-server/stations/repositories"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/general"
)

type (
	StationsController struct {
	}
)

func (sc StationsController) GETPath() string {
	return "/stations"
}

func NewStationsController() *StationsController {
	return &StationsController{}
}

func (sc *StationsController) GetStations(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	stations := make([]models.Station, 0)
	err := stationsrepository.Shared().GetStationsRepository().Find(nil).All(&stations)
	if err != nil {
		panic(err)
	}

	// Send response back
	response := stationsIO.GetStationsResponse{stations}
	basecontroller.SendResponse(w, response)
}