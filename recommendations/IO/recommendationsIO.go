package recommendationsIO

import "github.com/gordonseto/soundvis-server/stations/models"

type GetRecommendationsResponse struct {
	Recommendations []*models.Station	`json:"recommendations"`
}