package models

import "gopkg.in/mgo.v2/bson"

type (
	User struct {
		Id bson.ObjectId	`bson:"_id"`
		DeviceToken string 	`bson:"deviceToken"`
		IsPlaying bool		`bson:"isPlaying"`
		CurrentPlaying string	`bson:"currentPlaying"`
		Recordings []string	`bson:"recordings"`
		CreatedAt int64		`bson:"createdAt"`
	}
)