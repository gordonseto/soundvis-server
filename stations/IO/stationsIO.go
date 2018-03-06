package stationsIO

import "github.com/gordonseto/soundvis-server/stations/models"

type GetStationsResponse struct {
	Stations []models.Station	`json:"stations"`
}