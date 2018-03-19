package recordings

import (
	"github.com/gordonseto/soundvis-server/stations/models"
	"gopkg.in/mgo.v2/bson"
)

type Recording struct {
	Id string	`json:"id" bson:"_id"`
	Title string	`json:"title" bson:"title"`
	CreatorId string `json:"creatorId" bson:"creatorId"`
	StationId string	`json:"stationId" bson:"stationId"`	// this does not get sent to client
	Station *models.Station	`json:"station"`	// this is not saved in database
	StartDate int64	`json:"startDate" bson:"startDate"`
	EndDate int64	`json:"endDate" bson:"endDate"`
	RecordingURL string 	`json:"recordingUrl" bson:"recordingUrl"`
	Progress int64	`json:"progress" bson:"recording"`
	Status string	`json:"status" bson:"status"`
	JobId	string	`json:"jobId" bson:"jobId"`
	CreatedAt int64	`json:"createdAt" bson:"createdAt"`
	UpdatedAt int64	`json:"updatedAt" bson:"updatedAt"`
}

var RECORDING_ID_SUFFIX = "RSV"

var StatusFinished = "FINISHED"
var StatusPending = "PENDING"
var StatusInProgress = "IN_PROGRESS"
var StatusFailed = "FAILED"

// creates a recordingId
func CreateRecordingId() string {
	return bson.NewObjectId().Hex() + RECORDING_ID_SUFFIX
}

// returns true if id is a recordingId
func IdIsRecording(id string) bool {
	if len(id) < len(RECORDING_ID_SUFFIX) {
		return false
	}
	// RECORDING_ID_SUFFIX is at end of array, return true if last part of id == RECORDING_ID_SUFFIX
	return id[len(id)-len(RECORDING_ID_SUFFIX):] == RECORDING_ID_SUFFIX
}