package models

type (
	Station struct {
		Id string	`json:"id"`
		Name string	`json:"name"`
		DirbleId int `json:"dirbleId"`
		Country *Country	`json:"country"`
		StreamURL string `json:"streamUrl"`
		Genre	string	`json:"genre"`
		CreatedAt int64	`json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
	}

	Country struct {
		Code string 	`json:"code" bson:"country"`
		Name string		`json:"name" bson:"name"`
		Latitude float64	`json:"latitude" bson:"latitude"`
		Longitude float64	`json:"longitude" bson:"longitude"`
	}
)