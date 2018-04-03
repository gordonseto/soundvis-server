package streamhelper

import (
	"encoding/json"
	"errors"
	"github.com/gordonseto/soundvis-server/general"
	models2 "github.com/gordonseto/soundvis-server/recordings/models"
	"github.com/gordonseto/soundvis-server/recordings/repositories/recordingsrepository"
	"github.com/gordonseto/soundvis-server/recordings/repositories/recordingtracklistsrepository"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stations/repositories"
	"github.com/gordonseto/soundvis-server/stream/models"
	"log"
	"net/http"
	"github.com/gordonseto/soundvis-server/stationsfetcher"
	"github.com/gordonseto/soundvis-server/config"
)

// gets the audio stream url from currentPlaying
// currentPlaying is the id of a station or recording
func GetStreamURL(currentPlaying string, currentStation *models.Station) string {
	if currentPlaying == "" {
		return ""
	}
	if CurrentPlayingIsRecording(currentPlaying) {
		recording, err := recordingsrepository.Shared().FindRecordingById(currentPlaying)
		if err != nil {
			log.Println("Recording not found, currentPlaying - ", currentPlaying)
			return ""
		}
		return recording.RecordingURL
	} else {
		if currentStation != nil {
			return currentStation.StreamURL
		}
		return ""
	}
}

func CurrentPlayingIsRecording(currentPlaying string) bool {
	if currentPlaying == "" {
		return false
	}
	return models2.IdIsRecording(currentPlaying)
}

// gets the current station and the song from currentPlaying
// currentPlaying is an id for a station or recording
// progress is the progress in a recording, can be anything if currentPlaying is a station
func GetCurrentStationAndSongPlaying(currentPlaying string, progress int64) (*models.Station, *stream.Song, error) {
	if currentPlaying == "" {
		return nil, nil, nil
	}

	station, err := GetStation(currentPlaying)
	if err != nil {
		return nil, nil, err
	}
	if station != nil {
		song, err := GetCurrentSongPlaying(currentPlaying, progress, station)
		return station, song, err
	}
	return nil, nil, nil
}

// gets the station from currentPlaying
// currentPlaying is an id for a station or recording
func GetStation(currentPlaying string) (*models.Station, error) {
	if currentPlaying == "" {
		return nil, nil
	}

	if CurrentPlayingIsRecording(currentPlaying) {
		// if currentPlaying is recording, get recording and then its corresponding station
		recording, err := recordingsrepository.Shared().FindRecordingById(currentPlaying)
		if err != nil {
			return nil, err
		}
		if recording.Status == models2.StatusInProgress || recording.Status == models2.StatusPending {
			return nil, errors.New("Recording has not finished")
		}
		if recording.Status == models2.StatusFailed {
			return nil, errors.New("Recording is not valid")
		}
		stationId := recording.StationId
		if stationId == "" {
			return nil, errors.New("Recording has invalid stationId")
		}
		return stationsrepository.Shared().FindStationById(stationId)
	} else {
		return stationsrepository.Shared().FindStationById(currentPlaying)
	}
}

// gets the current song from currentPlaying
// currentPlaying is an id for a station or valid recording
// progress is the progress in a recording, can be anything if currentPlaying is a station
func GetCurrentSongPlaying(currentPlaying string, progress int64, station *models.Station) (*stream.Song, error) {
	if currentPlaying == "" {
		return nil, nil
	}

	if CurrentPlayingIsRecording(currentPlaying) {
		// get the tracklist for the recording
		tracklist, err := recordingtracklistsrepository.Shared().FindTracklistByRecordingId(currentPlaying)
		if err != nil {
			return nil, err
		}
		if len(tracklist.Tracklist) > 0 {
			song := tracklist.Tracklist[0].Song
			// iterate through the tracklist until progress < tracklist[i].Time
			for _, track := range tracklist.Tracklist {
				if track.Time <= progress {
					song = track.Song
				} else {
					break
				}
			}
			return song, nil
		}
		return nil, errors.New("Tracklist is empty")
	} else {
		if station.StreamURL == "" {
			return nil, errors.New("Station has no streamURL")
		}
		return stationsfetcher.GetCurrentSongPlayingShoutcast(station.StreamURL)
	}
}

// gets the imageURL for song
func GetImageURLForSong(song *stream.Song) error {
	if song == nil {
		return errors.New("Song is nil")
	}

	BASE_URL := "http://ws.audioscrobbler.com/2.0/?method=track.getInfo&api_key="
	url := BASE_URL + config.LAST_FM_API_KEY + "&artist=" + song.Name + "&track=" + song.Title + "&format=json"

	// make the request
	res, err := basecontroller.MakeRequest(url, http.MethodGet, 5)

	// parse the response
	type GetSongInfoResponse struct {
		Track struct {
			Album struct {
				Images []map[string]string `json:"image"`
			}	`json:"album"`
		}`json:"track"`
	}
	var response GetSongInfoResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return  err
	}

	// get the largest image in the response, which is the last one in the array
	if len(response.Track.Album.Images) > 0 {
		image := response.Track.Album.Images[len(response.Track.Album.Images)-1]
		if imageURL, ok := image["#text"]; ok {
			song.ImageURL = imageURL
			return nil
		}
	}
	return errors.New("Error parsing imageUrl")
}