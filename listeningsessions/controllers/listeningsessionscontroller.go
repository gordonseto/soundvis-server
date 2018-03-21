package listeningsessionscontroller

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/listeningsessions/repositories"
	"github.com/gordonseto/soundvis-server/listeningsessions/IO"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"sort"
	"github.com/gordonseto/soundvis-server/general"
)

type ListeningSessionsController struct {
}

func (lsc *ListeningSessionsController) GETPath() string {
	return "/stats"
}

func NewListeningSessionsController() *ListeningSessionsController {
	return &ListeningSessionsController{}
}

func (lsc *ListeningSessionsController) GetStats(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	sessions, err := listeningsessionsrepository.Shared().FindSessionsByUser(user.Id.Hex())
	if err != nil {
		panic(err)
	}

	stationDurationMap := make(map[string]*listeningsessionsIO.StationDuration)
	for _, s := range sessions {
		var stationDuration *listeningsessionsIO.StationDuration
		stationDuration, ok := stationDurationMap[s.StationId]
		if !ok {
			stationDuration = &listeningsessionsIO.StationDuration{
				Duration: 0,
			}
		}
		stationDuration.Duration += s.Duration
		stationDurationMap[s.StationId] = stationDuration
	}

	stationDurations := make([]*listeningsessionsIO.StationDuration, 0)
	for k, v := range stationDurationMap {
		v.Station, err = streamhelper.GetStation(k)
		if err != nil {
			panic(err)
		}
		stationDurations = append(stationDurations, v)
	}

	sort.Slice(stationDurations, func(i, j int) bool {
		return stationDurations[i].Duration > stationDurations[j].Duration
	})

	response := listeningsessionsIO.GetStatsResponse{}
	response.Rankings = stationDurations
	basecontroller.SendResponse(w, response)
}