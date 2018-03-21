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
	"sync"
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

	// get user's sessions
	sessions, err := listeningsessionsrepository.Shared().FindSessionsByUser(user.Id.Hex())
	if err != nil {
		panic(err)
	}

	// iterate through their sessions and aggregate total duration for each station
	stationDurationMap := make(map[string]*listeningsessionsIO.StationDuration)
	for _, session := range sessions {
		var stationDuration *listeningsessionsIO.StationDuration
		// get current duration from map
		stationDuration, ok := stationDurationMap[session.StationId]
		if !ok {
			// if not in map, create new entry
			stationDuration = &listeningsessionsIO.StationDuration{
				Duration: 0,
			}
		}
		// add current session's duration to total duration for station
		stationDuration.Duration += session.Duration
		stationDurationMap[session.StationId] = stationDuration
	}

	// for each stationDuration, get the full station object
	var waitGroup sync.WaitGroup
	for k, v := range stationDurationMap {
		waitGroup.Add(1)
		go getStationAsync(k, v, &waitGroup)
	}
	// wait for all calls to finish
	waitGroup.Wait()

	// iterate through map again and add entries into an array
	stationDurations := make([]*listeningsessionsIO.StationDuration, 0)
	for _, v := range stationDurationMap {
		stationDurations = append(stationDurations, v)
	}

	// sort array by decreasing duration
	sort.Slice(stationDurations, func(i, j int) bool {
		return stationDurations[i].Duration > stationDurations[j].Duration
	})

	response := listeningsessionsIO.GetStatsResponse{}
	response.Rankings = stationDurations
	basecontroller.SendResponse(w, response)
}

// gets station corresponding to stationId and sets it to stationDuration's station
func getStationAsync(stationId string, stationDuration *listeningsessionsIO.StationDuration, waitGroup *sync.WaitGroup) {
	station, err := streamhelper.GetStation(stationId)
	if err != nil {
		panic(err)
	}
	stationDuration.Station = station
	waitGroup.Done()
}