package models

import "gopkg.in/mgo.v2/bson"

type (
	Station struct {
		Id bson.ObjectId	`json:"id" bson:"_id"`
		Name string	`json:"name" bson:"name"`
		Country *Country	`json:"country" bson:"country"`
		StreamURL string `json:"streamUrl" bson:"streamUrl"`
		Genre	string	`json:"genre" bson:"genre"`
		CreatedAt int64	`json:"createdAt" bson:"createdAt"`
		UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
	}

	Country struct {
		Code string 	`json:"code" bson:"country"`
		Name string		`json:"name" bson:"name"`
		Latitude float64	`json:"latitude" bson:"latitude"`
		Longitude float64	`json:"longitude" bson:"longitude"`
	}
)