package recordingsIO

import (
	"github.com/gordonseto/soundvis-server/recordings/models"
)

type GetRecordingsResponse struct {
	Recordings []*recordings.Recording `json:"recordings"`
}

type CreateRecordingRequest struct {
	StationId string	`json:"stationId"`
	Title string 	`json:"title"`
	StartDate int64	`json:"startDate"`
	EndDate int64 	`json:"endDate"`
}

type CreateRecordingResponse struct {
	Recording *recordings.Recording 	`json:"recording"`
}

type UpdateRecordingRequest struct {
	RecordingId string	`json:"recordingId"`
	StationId string	`json:"stationId"`
	Title string 	`json:"title"`
	StartDate int64	`json:"startDate"`
	EndDate int64 	`json:"endDate"`
}

type DeleteRecordingRequest struct {
	RecordingId string `json:"recordingId"`
}

type DeleteRecordingResponse struct {
	Ok bool	`json:"ok"`
}