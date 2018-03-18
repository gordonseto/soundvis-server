package streamIO

import (
	"github.com/gordonseto/soundvis-server/stations/models"
	"github.com/gordonseto/soundvis-server/stream/models"
)

type GetCurrentStreamResponse struct {
	IsPlaying bool 	`json:"isPlaying"`
	CurrentPlaying string `json:"currentPlaying"`
	CurrentStation *models.Station	`json:"currentStation"`
	CurrentStreamURL string	`json:"currentStreamUrl"`
	CurrentSong *stream.Song `json:"currentSong"`
}

type SetCurrentStreamRequest struct {
	IsPlaying bool `json:"isPlaying"`
	CurrentStream string	`json:"currentStream"`
}