package authentication

import (
	"net/http"
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/users/models"
	"errors"
	"github.com/gordonseto/soundvis-server/users/controllers"
)

var HEADER_USER_ID_KEY = "userId"

func CheckAuthentication(r *http.Request, session *mgo.Session) (*models.User, error) {
	userId := r.Header.Get(HEADER_USER_ID_KEY)
	if (userId == "") {
		err := errors.New("User not authorized")
		return nil, err
	}
	user, err := users.FindUserById(session, userId)
	return user, err
}