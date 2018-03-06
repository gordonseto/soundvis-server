package users

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/users/IO"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/config"
	"github.com/gordonseto/soundvis-server/users/models"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type (
	UsersController struct{
		session *mgo.Session
	}
)

func NewUsersController(s *mgo.Session) *UsersController {
	return &UsersController{s}
}

func (uc UsersController) getCollectionName() string {
	return "users"
}

func (uc UsersController) getCollection() *mgo.Collection {
	return uc.session.DB(config.DB_NAME).C(uc.getCollectionName())
}

func (uc UsersController) POSTPath() string {
	return "/users"
}

func (uc UsersController) CreateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// get deviceToken from request
	request := usersIO.CreateUserRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}
	// if no deviceToken, return badRequest
	if request.DeviceToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid deviceToken")
		return
	}

	// find user in database, if already contained, just return user
	user := models.User{}
	if err := findUserByDeviceToken(uc, request.DeviceToken, &user); err != nil {
		// no user found, create user
		user.Id = bson.NewObjectId()
		user.DeviceToken = request.DeviceToken
		user.CreatedAt = time.Now().Unix()
		// insert into collection
		if err = uc.getCollection().Insert(user); err != nil {
			panic(err)
		}
	}

	response := usersIO.CreateUserResponse{}
	response.User = user

	responseJSON, err := json.Marshal(&response)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", responseJSON)
}

func findUserByDeviceToken(uc UsersController, deviceToken string, user *models.User) error {
	return uc.getCollection().Find(bson.M{"deviceToken":deviceToken}).One(&user)
}