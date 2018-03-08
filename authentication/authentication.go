package authentication

import (
	"net/http"
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/users/models"
	"errors"
	"github.com/gordonseto/soundvis-server/users/controllers"
)

var HEADER_USER_ID_KEY = "userId"
var HEADER_USER_AGENT_KEY = "User-Agent"

func CheckAuthentication(r *http.Request, session *mgo.Session) (*models.User, error) {
	userId := r.Header.Get(HEADER_USER_ID_KEY)
	if (userId == "") {
		err := errors.New("User not authorized")
		return nil, err
	}
	user, err := users.FindUserById(session, userId)
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