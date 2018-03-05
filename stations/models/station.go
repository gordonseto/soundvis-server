package models

type (
	Station struct {
		Id string	`json:"id"`
		Name string	`json:"name"`
		DirbleId int `json:"dirbleId"`
		Country string `json:"country"`
		StreamURL string `json:"streamUrl"`
		Genre	string	`json:"genre"`
		CreatedAt int64	`json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
	}
)