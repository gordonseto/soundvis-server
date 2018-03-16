package models

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/gordonseto/soundvis-server/stations/models"
)

type Recording struct {
	Id bson.ObjectId	`json:"id" bson:"_id"`
	Title string	`json:"title" bson:"title"`
	CreatorId string `json:"creatorId" bson:"creatorId"`
	Station *models.Station	`json:"station" bson:"station"`
	StartDate int64	`json:"startDate" bson:"startDate"`
	EndDate int64	`json:"endDate" bson:"endDate"`
	RecordingURL string 	`json:"recordingUrl" bson:"recordingUrl"`
	Progress int64	`json:"progress" bson:"recording"`
	CreatedAt int64	`json:"createdAt" bson:"createdAt"`
	UpdatedAt int64	`json:"updatedAt" bson:"updatedAt"`
}