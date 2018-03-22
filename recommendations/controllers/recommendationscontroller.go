package recommendationscontroller

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/authentication"
	"os/exec"
	"strings"
	"github.com/gordonseto/soundvis-server/recommendations/IO"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/general"
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
	cmd := exec.Command("python",  "recommendations/recommender.py", user.Id.Hex())
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	// output is a comma delimited string of stationIds, split into array
	recommendations := strings.Split(string(out), ",")

	// get top 5 recommendations' stations
	stations := make([]*models.Station, 0)
	for _, recommendation := range recommendations[:5] {
		station, err := streamhelper.GetStation(recommendation)
		if err != nil {
			panic(err)
		}
		stations = append(stations, station)
	}

	response := recommendationsIO.GetRecommendationsResponse{}
	response.Recommendations = stations
	basecontroller.SendResponse(w, response)
}