package dirblestationsfetcher

import (
	"encoding/json"
	"github.com/gordonseto/soundvis-server/config"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/stream/models"
	"gopkg.in/mgo.v2"
	"net/http"
	"github.com/gordonseto/soundvis-server/stations/models"
	"time"
	"strconv"
)

// Gets stations from dirble and converts to Stations
func GetStations(dirbleStationId string, perPage, offset int, session *mgo.Session) ([]models.Station, error) {
	// get dirbleStations from their API
	dirbleStations, err := getDirbleStations(dirbleStationId, perPage, offset)
	if err != nil {
		return nil, err
	}

	stations := make([]models.Station, len(dirbleStations))

	// get mapping of countries
	var countriesMap map[string]models.Country
	if len(stations) > 0 {
		countriesMap = getAllCountries(session)
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
func getAllCountries(session *mgo.Session) map[string]models.Country {
	var countries []models.Country
	err := session.DB(config.DB_NAME).C("countries").Find(nil).All(&countries)
	if err != nil {
		panic(err)
	}

	var countriesMap = make(map[string]models.Country)
	for _, country := range countries {
		countriesMap[country.Code] = country
	}
	return countriesMap
}

// takes in currentPlaying and returns the station corresponding with that id
// currentPlaying is an id of a dirble station
func GetStationDirble(dirbleId string, session *mgo.Session) (*models.Station, error) {
	if dirbleId == "" {
		return nil, nil
	}

	stations, err := GetStations(dirbleId, 0, 0, session)
	if err != nil || len(stations) < 1 {
		return nil, err
	}
	return &stations[0], nil
}

// gets the most recent playing song in currentPlaying
// currentPlaying is the id of a dirble station
func GetCurrentSongPlayingDirble(dirbleId string) (*stream.Song, error) {
	if dirbleId == "" {
		return nil, nil
	}

	// build url
	url := "http://api.dirble.com/v2/station/" + dirbleId + "/song_history?token=" + config.DIRBLE_API_KEY
	songs := make([]stream.Song, 0)

	// make request
	body, err := basecontroller.MakeRequest(url, http.MethodGet, 10)
	if err != nil {
		return nil, err
	}

	// parse request
	err = json.Unmarshal(body, &songs)
	if err != nil {
		return nil, err
	}

	if len(songs) < 1 {
		return nil, nil
	}

	return &songs[0], nil
}
