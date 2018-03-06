package users

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/users/IO"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
)

type (
	UsersController struct{
		session *mgo.Session
	}
)

func NewUsersController(s *mgo.Session) *UsersController {
	return &UsersController{s}
}

func (uc UsersController) POSTPath() string {
	return "/users"
}

func (uc UsersController) CreateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	request := usersIO.CreateUserRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}
	if request.DeviceToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid deviceToken")
		return
	}


	w.WriteHeader(200)
}