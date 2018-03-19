package streamhelper

import (
	"github.com/gordonseto/soundvis-server/stationsfetcher"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stream/models"
	"github.com/gordonseto/soundvis-server/stations/repositories"
	models2 "github.com/gordonseto/soundvis-server/recordings/models"
	"github.com/gordonseto/soundvis-server/recordings/repositories"
	"errors"
	"log"
)

// gets the audio stream url from currentPlaying
// currentPlaying is the id of a station or recording
func GetStreamURL(currentPlaying string, currentStation *models.Station) string {
	if currentPlaying == "" {
		return ""
	}
	if currentPlayingIsRecording(currentPlaying) {
		recording, err := recordingsrepository.Shared().FindRecordingById(currentPlaying)
		if err != nil {
			log.Println("Recording not found, currentPlaying - ", currentPlaying)
			return ""
		}
		return recording.RecordingURL
	} else {
		return currentStation.StreamURL
	}
}

func currentPlayingIsRecording(currentPlaying string) bool {
	if currentPlaying == "" {
		return false
	}
	return models2.IdIsRecording(currentPlaying)
}

// gets the current station and the song from currentPlaying
// currentPlaying is an id for a station or recording
func GetCurrentStationAndSongPlaying(currentPlaying string) (*models.Station, *stream.Song, error) {
	if currentPlaying == "" {
		return nil, nil, nil
	}

	station, err := GetStation(currentPlaying)
	if err != nil {
		return nil, nil, err
	}
	song, err := GetCurrentSongPlaying(currentPlaying, station)
	return station, song, err
}

// gets the station from currentPlaying
// currentPlaying is an id for a station or recording
func GetStation(currentPlaying string) (*models.Station, error) {
	if currentPlaying == "" {
		return nil, nil
	}

	if currentPlayingIsRecording(currentPlaying) {
		// if currentPlaying is recording, get recording and then its corresponding station
		recording, err := recordingsrepository.Shared().FindRecordingById(currentPlaying)
		if err != nil {
			return nil, err
		}
		if recording.RecordingURL == "" {
			return nil, errors.New("Recording has not finished")
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
// currentPlaying is an id for a station or recording
func GetCurrentSongPlaying(currentPlaying string, station *models.Station) (*stream.Song, error) {
	if currentPlaying == "" {
		return nil, nil
	}

	if currentPlayingIsRecording(currentPlaying) {
		// TODO: Implement this
		return nil, nil
	} else {
		return stationsfetcher.GetCurrentSongPlayingShoutcast(station.StreamURL)
	}
}