package listeningsessionsIO

import "github.com/gordonseto/soundvis-server/stations/models"

type GetStatsResponse struct {
	Rankings []*StationDuration	`json:"rankings"`
}

type StationDuration struct {
	Duration int64	`json:"duration"`
	Station *models.Station	`json:"station"`
}