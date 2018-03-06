package stations

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/config"
	"strconv"
	"github.com/gordonseto/soundvis-server/stations/IO"
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
)

type (
	StationsController struct {
		session *mgo.Session
	}
)

func (sc StationsController) GETPath() string {
	return "/stations"
}

func NewStationsController(s *mgo.Session) *StationsController {
	return &StationsController{s}
}

func (sc StationsController) getCountriesCollection() *mgo.Collection {
	return sc.session.DB(config.DB_NAME).C("countries")
}

func (sc StationsController) GetStations(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// get dirbleStations from their API
	dirbleStations, err := getDirbleStations(20, 0)
	if err != nil {
		panic(err)
	}

	// convert dirbleStations to stations
	stations := make([]models.Station, len(dirbleStations))

	// get mapping of countries
	var countriesMap map[string]models.Country
	if len(stations) > 0 {
		countriesMap = sc.getAllCountries()
	}

	// iterate through dirbleStations and create stations
	for i, dirbleStation := range dirbleStations {
		station := models.Station{
			Name: dirbleStation.Name,
			DirbleId: dirbleStation.Id,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		// get first streamUrl in dirbleStation array of streamUrls
		if len(dirbleStation.Streams) > 0 {
			station.StreamURL = dirbleStation.Streams[0].Stream
		}
		// get first category in dirbleStation categories
		if len(dirbleStation.Categories) > 0 {
			station.Genre = dirbleStation.Categories[0].Title
		}
		// get country object from countriesMap
		country := countriesMap[dirbleStation.Country]
		station.Country = &country

		stations[i] = station
	}

	// Send response back
	response := stationsIO.GetStationsResponse{stations}
	responseJSON, err := json.Marshal(&response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", responseJSON)
}

// data structures used to parse dirbleStation request
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

// gets perPage popular stations from dirble
func getDirbleStations(perPage, offset int) ([]dirbleStation, error) {
	url := "http://api.dirble.com/v2/stations/popular?token=" + config.DIRBLE_API_KEY + "&per_page=" + strconv.Itoa(perPage) + "&offset=" + strconv.Itoa(offset)

	// Make request
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

	// Read response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Create DirbleStations from response
	dirbleStations := make([]dirbleStation, 0)
	err = json.Unmarshal(body, &dirbleStations)
	if err != nil {
		return nil, err
	}

	return dirbleStations, nil
}

// gets all countries and returns map where key = country code, value = country
func (sc StationsController) getAllCountries() map[string]models.Country {
	var countries []models.Country
	err := sc.getCountriesCollection().Find(nil).All(&countries)
	if err != nil {
		panic(err)
	}

	var countriesMap = make(map[string]models.Country)
	for _, country := range countries {
		countriesMap[country.Code] = country
	}
	return countriesMap
}