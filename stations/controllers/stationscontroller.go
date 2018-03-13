package stations

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"time"
	"encoding/json"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/config"
	"strconv"
	"github.com/gordonseto/soundvis-server/stations/IO"
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/general"
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

func (sc *StationsController) getCountriesCollection() *mgo.Collection {
	return sc.session.DB(config.DB_NAME).C("countries")
}

func (sc *StationsController) GetStations(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	stations, err := GetStations(sc, "", 20, 0)
	if err != nil {
		panic(err)
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


// Gets stations from dirble and converts to Stations
func GetStations(sc *StationsController, dirbleStationId string, perPage, offset int) ([]models.Station, error) {
	// get dirbleStations from their API
	dirbleStations, err := getDirbleStations(dirbleStationId, perPage, offset)
	if err != nil {
		return nil, err
	}

	stations := make([]models.Station, len(dirbleStations))

	// get mapping of countries
	var countriesMap map[string]models.Country
	if len(stations) > 0 {
		countriesMap = sc.getAllCountries()
	}

	// iterate through dirbleStations and create stations
	for i, dirbleStation := range dirbleStations {
		station := dirbleStationToStation(&dirbleStation, countriesMap)
		stations[i] = *station
	}

	return stations, nil
}

func dirbleStationToStation(dirbleStation *dirbleStation, countriesMap map[string]models.Country) *models.Station {
	station := models.Station{
		Name: dirbleStation.Name,
		//DirbleId: dirbleStation.Id,
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
	return &station
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

// gets perPage popular stations from dirble if dirbleStationId is empty, else gets that specific dirbleStation
func getDirbleStations(dirbleStationId string, perPage, offset int) ([]dirbleStation, error) {
	// if dirbleStationId is present, hit endpoint for specific station, else hit popular stations
	var url string
	if dirbleStationId == "" {
		url = "http://api.dirble.com/v2/stations/popular?token=" + config.DIRBLE_API_KEY + "&per_page=" + strconv.Itoa(perPage) + "&offset=" + strconv.Itoa(offset)
	} else {
		url = "http://api.dirble.com/v2/station/" + dirbleStationId + "?token=" + config.DIRBLE_API_KEY
	}

	// Make request
	body, err := basecontroller.MakeRequest(url, http.MethodGet, 10)
	if err != nil {
		return nil, err
	}

	// Create DirbleStations from response
	dirbleStations := make([]dirbleStation, 0)

	// if finding all dirbleStations, response will be an array
	if dirbleStationId == "" {
		err = json.Unmarshal(body, &dirbleStations)
		if err != nil {
			return nil, err
		}
	} else {
		// else it is a single dirbleStation, parse and append to array as only value
		var dirbleStation dirbleStation
		err = json.Unmarshal(body, &dirbleStation)
		if err != nil {
			return nil, err
		}
		dirbleStations = append(dirbleStations, dirbleStation)
	}

	return dirbleStations, nil
}

// gets all countries and returns map where key = country code, value = country
func (sc *StationsController) getAllCountries() map[string]models.Country {
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