package stationsfetcher

import (
	"github.com/gordonseto/soundvis-server/general"
	"net/http"
	"encoding/xml"
	"fmt"
	"strings"
	"github.com/gordonseto/soundvis-server/stations/models"
	"time"
	"sync"
	"errors"
)

type ShoutCastStationResponse struct {
	XMLName	xml.Name	`xml:"stationlist"`
	StationList []ShoutcastStation	`xml:"station"`
}

type ShoutcastStation struct {
	XMLName xml.Name	`xml:"station"`
	Id string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	Genre string `xml:"genre,attr"`
}

type TuneInResponse struct {
	XMLName xml.Name	`xml:"playlist"`
	Tracklist	TrackList	`xml:"trackList"`
}

type TrackList struct {
	XMLName xml.Name	`xml:"trackList"`
	Tracks	[]Track		`xml:"track"`
}

type Track struct {
	XMLName xml.Name	`xml:"track"`
	Location string		`xml:"location"`
}

type html struct {
	Body body `xml:"body"`
}
type body struct {
	Content string `xml:",innerxml"`
}

var waitGroup sync.WaitGroup
var mutex = &sync.Mutex{}

func FetchAndStoreStations() {
	FETCH_STATIONS_URL := "http://api.shoutcast.com/legacy/Top500?k=t7kGHgPoxtvtuUOc"

	stations := make([]models.Station, 0)

	// Fetch all the stations
	body, err := basecontroller.MakeRequest(FETCH_STATIONS_URL, http.MethodGet, 10)
	if err != nil {
		fmt.Println(err)
	}
	var response ShoutCastStationResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
	}

	// iterate through and get info for each station
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
	}
}

// takes a shoutcastStation and fetches its info to create a models.Station
// if unable to, returns nil for station
func getStationInfo(shoutcastStation ShoutcastStation) (*models.Station, error) {
	defer fmt.Println("Done station!")

	TUNE_IN_URL := "http://yp.shoutcast.com/sbin/tunein-station.xspf?id="

	fmt.Println(shoutcastStation.Name + " " + shoutcastStation.Genre + " " + shoutcastStation.Id)

	// Make request for streamURL
	body, err := basecontroller.MakeRequest(TUNE_IN_URL + shoutcastStation.Id , http.MethodGet, 2)
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
		track := tuneInResponse.Tracklist.Tracks[0]
		fmt.Println(track.Location)
		// remove trailing url component
		streamBaseArray := strings.Split(track.Location, "/")
		streamBase := ""
		for j := 0; j < len(streamBaseArray)-1; j++ {
			streamBase += streamBaseArray[j] + "/"
		}
		fmt.Println(streamBase)
		// send request to urlbase + "7"
		// if this request succeeds, use this station, else discard it
		res, err := basecontroller.MakeRequest(streamBase+"7", http.MethodGet, 5)
		if err != nil {
			return nil, err
		} else {
			fmt.Println(res)
			h := html{}
			err := xml.NewDecoder(strings.NewReader(string(res))).Decode(&h)
			if err != nil {
				return nil, err
			} else {
				fmt.Println(h.Body.Content)
				station := models.Station{
					Name:      shoutcastStation.Name,
					Genre:     shoutcastStation.Genre,
					StreamURL: track.Location,
					CreatedAt: time.Now().Unix(),
					UpdatedAt: time.Now().Unix(),
				}
				// TODO: Remove this
				station.Id = h.Body.Content
				return &station, nil
			}
		}
	}
	return nil, errors.New("Station has no tracks in tracklist")
}