package listeningsession

import "gopkg.in/mgo.v2/bson"

type ListeningSession struct {
	Id	bson.ObjectId	`bson:"_id"`
	UserId	string	`bson:"userId"`
	StationId string	`bson:"stationId"`
	Duration int64	`bson:"duration"`
	StartTime int64	`bson:"startTime"`
}