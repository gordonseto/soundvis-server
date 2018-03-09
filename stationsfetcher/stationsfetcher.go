package stationsfetcher

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stream/controllers"
	"net/http"
	"sync"
	"time"
	"encoding/json"
	"strings"
)

type ShoutCastStationResponse struct {
	XMLName     xml.Name           `xml:"stationlist"`
	StationList []ShoutcastStation `xml:"station"`
}

type ShoutcastStation struct {
	XMLName xml.Name `xml:"station"`
	Id      string   `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
	Genre   string   `xml:"genre,attr"`
}

type TuneInResponse struct {
	XMLName   xml.Name  `xml:"playlist"`
	Tracklist TrackList `xml:"trackList"`
}

type TrackList struct {
	XMLName xml.Name `xml:"trackList"`
	Tracks  []Track  `xml:"track"`
}

type Track struct {
	XMLName  xml.Name `xml:"track"`
	Location string   `xml:"location"`
}

var waitGroup sync.WaitGroup
var mutex = &sync.Mutex{}

func FetchAndStoreStations() []models.Station {
	FETCH_STATIONS_URL := "http://api.shoutcast.com/legacy/Top500?k=t7kGHgPoxtvtuUOc"

	stations := make([]models.Station, 0)

	// Fetch all the stations
	body, err := basecontroller.MakeRequest(FETCH_STATIONS_URL, http.MethodGet, 10)
	if err != nil {
		fmt.Println(err)
	}
	// Parse response
	var response ShoutCastStationResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
	}

	// iterate through each shoutcast station in the response and get info for each station
	for i := 0; i < 100; i++ {
		shoutcastStation := response.StationList[i]
		// spin own routine for each station
		go func() {
			// increment waitgroup
			waitGroup.Add(1)
			// get info for station
			station, err := getStationInfo(shoutcastStation)
			if err != nil {
				fmt.Println(err)
			} else if station != nil {
				// if station not nil, add to stations
				mutex.Lock()
				stations = append(stations, *station)
				mutex.Unlock()
			}
			// decrement waitgroup
			waitGroup.Done()
		}()
	}

	// wait for all station goroutines to finish
	waitGroup.Wait()

	for _, station := range stations {
		fmt.Println(station)
		fmt.Println(station.Country)
	}

	return stations
}

// takes a shoutcastStation and fetches its info to create a models.Station
// if unable to, returns nil for station
func getStationInfo(shoutcastStation ShoutcastStation) (*models.Station, error) {
	defer fmt.Println("Done station!")

	TUNE_IN_URL := "http://yp.shoutcast.com/sbin/tunein-station.xspf?id="

	fmt.Println(shoutcastStation.Name + " " + shoutcastStation.Genre + " " + shoutcastStation.Id)

	// Make request for streamURL
	body, err := basecontroller.MakeRequest(TUNE_IN_URL + shoutcastStation.Id, http.MethodGet, 2)
	if err != nil {
		return nil, err
	}
	// parse response
	var tuneInResponse TuneInResponse
	err = xml.Unmarshal(body, &tuneInResponse)
	if err != nil {
		return nil, err
	}

	// Get first streamURL out of array
	if len(tuneInResponse.Tracklist.Tracks) > 0 {
		streamURL := tuneInResponse.Tracklist.Tracks[0].Location

		// get currentSong, this is to test if stream should be used
		currentSong, err := stream.GetCurrentSongPlayingShoutcast(streamURL)

		// if there is an error, discard station
		if err != nil {
			return nil, err
		} else {
			// get country for station
			domain := getBaseDomain(streamURL)
			if domain != "" {
				country, err := getCountryForAddress(domain)
				if err != nil {
					return nil, err
				}

				station := models.Station{
					Name:      shoutcastStation.Name,
					Genre:     shoutcastStation.Genre,
					StreamURL: streamURL,
					Country: country,
					CreatedAt: time.Now().Unix(),
					UpdatedAt: time.Now().Unix(),
				}
				// TODO: Remove this
				station.Id = currentSong
				return &station, nil
			} else {
				return nil, errors.New("Station has invalid streamURL")
			}
		}

	}
	return nil, errors.New("Station has no tracks in tracklist")
}

type CountryResponse struct {
	Country string `json:"country"`
	CountryCode string `json:"countryCode"`
	Lat float64	`json:"lat"`
	Lon float64 `json:lon`
}

func getBaseDomain(url string) string {
	domainArray := strings.Split(url, "/")
	if len(domainArray) >= 3 {
		domain := strings.Split(domainArray[2], ":")[0]
		return domain
	}
	return ""
}

func getCountryForAddress(address string) (*models.Country, error) {
	res, err := basecontroller.MakeRequest("http://ip-api.com/json/" + address, http.MethodGet, 10)
	if err != nil {
		return nil, err
	}
	var countryResponse CountryResponse
	err = json.Unmarshal(res, &countryResponse)
	if err != nil {
		return nil, err
	}
	country := models.Country{
		Name: countryResponse.Country,
		Code: countryResponse.CountryCode,
		Latitude: countryResponse.Lat,
		Longitude: countryResponse.Lon,
	}
	return &country, nil
}