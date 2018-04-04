package models

import "gopkg.in/mgo.v2/bson"

type (
	User struct {
		Id bson.ObjectId	`bson:"_id"`
		DeviceToken string 	`bson:"deviceToken"`
		IsPlaying bool		`bson:"isPlaying"`
		CurrentPlaying string	`bson:"currentPlaying"`	// this is a stationId or recordingId
		CurrentVolume int	`bson:"currentVolume"`
		Recordings []string	`bson:"recordings"`
		CreatedAt int64		`bson:"createdAt"`
		StreamUpdatedAt int64 `bson:"streamUpdatedAt"`
	}
)