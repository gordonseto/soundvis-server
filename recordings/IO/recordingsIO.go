package recordingsIO

import (
	"github.com/gordonseto/soundvis-server/recordings/models"
)

type GetRecordingsResponse struct {
	Recordings []*models.Recording `json:"recordings"`
}

type CreateRecordingRequest struct {
	StationId string	`json:"stationId"`
	Title string 	`json:"title"`
	StartDate int64	`json:"startDate"`
	EndDate int64 	`json:"endDate"`
}

type CreateRecordingResponse struct {
	Recording *models.Recording 	`json:"recording"`
}