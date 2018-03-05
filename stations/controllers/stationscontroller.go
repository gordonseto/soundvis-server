package stations

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"github.com/gordonseto/soundvis-server/stations/models"
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

	// get dirbleStations from their API
	dirbleStations, err := getDirbleStations()
	if handleError(w, err) {
		return
	}

	// convert dirbleStations to stations
	stations := make([]models.Station, len(dirbleStations))
	for i, dirbleStation := range dirbleStations {
		station := models.Station{
			Name: dirbleStation.Name,
			DirbleId: dirbleStation.Id,
			Country: dirbleStation.Country,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		if len(dirbleStation.Streams) > 0 {
			station.StreamURL = dirbleStation.Streams[0].Stream
		}
		if len(dirbleStation.Categories) > 0 {
			station.Genre = dirbleStation.Categories[0].Title
		}
		stations[i] = station
	}

	stationsJSON, err := json.Marshal(stations)
	if handleError(w, err) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", stationsJSON)
}

type dirbleStation struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Country string `json:"country"`
	Categories []dirbleCategory `json:"categories"`
	Streams []dirbleStream	`json:"streams"`
}

type dirbleCategory struct {
	Title string `json:"title"`
}

type dirbleStream struct {
	Stream string `json:"stream"`
}

func getDirbleStations() ([]dirbleStation, error) {
	url := "http://api.dirble.com/v2/stations/popular?token=1aa6f199daa8d021c6c992800b&per_page=10"

	dirbleClient := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := dirbleClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	dirbleStations := make([]dirbleStation, 0)
	err = json.Unmarshal(body, &dirbleStations)
	if err != nil {
		return nil, err
	}

	return dirbleStations, nil
}

func handleError(w http.ResponseWriter, err error) bool {
	if err != nil {
		fmt.Println(err)
		return true
	}
	return false
}