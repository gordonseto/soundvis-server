package authentication

import (
	"net/http"
	"github.com/gordonseto/soundvis-server/users/models"
	"errors"
	"github.com/gordonseto/soundvis-server/users/repositories"
)

var HEADER_USER_ID_KEY = "userId"
var HEADER_USER_AGENT_KEY = "User-Agent"

func CheckAuthentication(r *http.Request) (*models.User, error) {
	userId := r.Header.Get(HEADER_USER_ID_KEY)
	if (userId == "") {
		err := errors.New("User not authorized")
		return nil, err
	}
	return FindUser(userId)
}

func FindUser(userId string) (*models.User, error) {
	user, err := usersrepository.Shared().FindUserById(userId)
	if err != nil {
		return nil, errors.New("User does not exist")
	}
	return user, err
}

const (
	DE1 = 1
	ANDROID = 2
)

func GetUserAgent(r *http.Request) (int, error) {
	userAgent := r.Header.Get(HEADER_USER_AGENT_KEY)

	if userAgent == "de1" {
		return DE1, nil
	} else if userAgent == "android" {
		return ANDROID, nil
	}
	err := errors.New("Invalid User Agent")
	return 0, err
}