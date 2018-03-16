package streammanager

import (
	"github.com/gordonseto/soundvis-server/stationsfetcher"
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stations/repositories"
	"github.com/gordonseto/soundvis-server/stream/models"
)

type StreamManager struct {
	stationsRepository *stationsrepository.StationsRepository
}

func NewStreamManager(sr *stationsrepository.StationsRepository) *StreamManager {
	return &StreamManager{sr}
}

// gets the audio stream url from currentPlaying
// currentPlaying is the id of a station or recording
func (sm *StreamManager) GetStreamURL(currentPlaying string, currentStation *models.Station) string {
	if currentPlaying == "" {
		return ""
	}
	if currentPlayingIsRecording(currentPlaying) {
		// TODO: Implement this
		return ""
	} else {
		return currentStation.StreamURL
	}
}

func currentPlayingIsRecording(currentPlaying string) bool {
	if currentPlaying == "" {
		return false
	}
	return false
}

// gets the current station and the song from currentPlaying
// currentPlaying is an id for a station or recording
func (sm *StreamManager) GetCurrentStationAndSongPlaying(currentPlaying string) (*models.Station, *stream.Song, error) {
	if currentPlaying == "" {
		return nil, nil, nil
	}

	station, err := sm.GetStation(currentPlaying)
	if err != nil {
		return nil, nil, err
	}
	song, err := sm.GetCurrentSongPlaying(currentPlaying, station)
	return station, song, err
}

// gets the station from currentPlaying
// currentPlaying is an id for a station or recording
func (sm *StreamManager) GetStation(currentPlaying string) (*models.Station, error) {
	if currentPlaying == "" {
		return nil, nil
	}

	if currentPlayingIsRecording(currentPlaying) {
		// TODO: Implement this
		return nil, nil
	} else {
		return sm.stationsRepository.FindStationById(currentPlaying)
	}
}

// gets the current song from currentPlaying
// currentPlaying is an id for a station or recording
func (sm *StreamManager) GetCurrentSongPlaying(currentPlaying string, station *models.Station) (*stream.Song, error) {
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