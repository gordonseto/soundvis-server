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
	"errors"
)

type (
	UsersController struct{
		session *mgo.Session
	}
)

func NewUsersController(s *mgo.Session) *UsersController {
	return &UsersController{s}
}

func getCollectionName() string {
	return "users"
}

func getCollection(session *mgo.Session) *mgo.Collection {
	return session.DB(config.DB_NAME).C(getCollectionName())
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
	if err := FindUserByDeviceToken(uc.session, request.DeviceToken, &user); err != nil {
		// no user found, create user
		user.Id = bson.NewObjectId()
		user.DeviceToken = request.DeviceToken
		user.CreatedAt = time.Now().Unix()
		// insert into collection
		if err = getCollection(uc.session).Insert(user); err != nil {
			panic(err)
		}
		// find user in collection
		if err = FindUserByDeviceToken(uc.session, request.DeviceToken, &user); err != nil {
			// if not found this time, there is an error
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

func FindUserByDeviceToken(session *mgo.Session, deviceToken string, user *models.User) error {
	return getCollection(session).Find(bson.M{"deviceToken":deviceToken}).One(&user)
}

func FindUserById(session *mgo.Session, userId string) (*models.User, error) {
	var user models.User
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("Invalid userId")
	}

	oid := bson.ObjectIdHex(userId)

	err := getCollection(session).FindId(oid).One(&user)
	return &user, err
}