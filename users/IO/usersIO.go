package usersIO

import (
	"github.com/gordonseto/soundvis-server/users/models"
)

type CreateUserRequest struct {
	DeviceToken string	`json:"deviceToken"`
}

type CreateUserResponse struct {
	User models.User	`json:"user"`
}