package stations

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"fmt"
)

type (
	StationsController struct {}
)

func (sc StationsController) GETPath() string {
	return "/stations"
}

func NewStationsController() *StationsController {
	return &StationsController{}
}

func (sc StationsController) GetStations(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "Hello World!")
}