package recommendationscontroller

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/authentication"
	"strings"
	"github.com/gordonseto/soundvis-server/recommendations/IO"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/general"
	"sync"
	"github.com/gordonseto/soundvis-server/scripthelper"
)

type RecommendationsController struct {
}

func NewRecommendationsController() *RecommendationsController {
	return &RecommendationsController{}
}

func (rc *RecommendationsController) GETPath() string {
	return "/recommendations"
}

func (rc *RecommendationsController) GetRecommendations(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	// run python script to get sorted list of station recommendations
	out, err := scripthelper.RunRecommenderScript(user.Id.Hex())
	if err != nil {
		panic(err)
	}
	// output is a comma delimited string of stationIds, split into array
	recommendations := strings.Split(string(out), ",")

	// get top 5 recommendations' stations
	stationsMap := make(map[string]*models.Station)
	var waitGroup sync.WaitGroup
	for _, recommendation := range recommendations[:5] {
		waitGroup.Add(1)
		go FetchStationAndPopulateMap(recommendation, stationsMap, &waitGroup)
	}
	waitGroup.Wait()

	// fill response with stations
	stations := make([]*models.Station, 0)
	for _, recommendation := range recommendations[:5] {
		if station, ok := stationsMap[recommendation]; ok {
			stations = append(stations, station)
		}
	}
	response := recommendationsIO.GetRecommendationsResponse{}
	response.Recommendations = stations
	basecontroller.SendResponse(w, response)
}

func FetchStationAndPopulateMap(stationId string, stations map[string]*models.Station, waitGroup *sync.WaitGroup) {
	station, err := streamhelper.GetStation(stationId)
	if err != nil {
		panic(err)
	}
	stations[stationId] = station
	waitGroup.Done()
}