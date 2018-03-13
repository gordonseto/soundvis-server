package users

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/gordonseto/soundvis-server/users/IO"
	"encoding/json"
	"fmt"
	"github.com/gordonseto/soundvis-server/users/models"
	"gopkg.in/mgo.v2/bson"
	"time"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/users/repositories"
)

type (
	UsersController struct{
		usersRepository *usersrepository.UsersRepository
	}
)

func NewUsersController(ur *usersrepository.UsersRepository) *UsersController {
	return &UsersController{ur}
}

func (uc *UsersController) POSTPath() string {
	return "/users"
}

func (uc *UsersController) CreateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
	if err := uc.usersRepository.FindUserByDeviceToken(request.DeviceToken, &user); err != nil {
		// no user found, create user
		user.Id = bson.NewObjectId()
		user.DeviceToken = request.DeviceToken
		user.CreatedAt = time.Now().Unix()
		// insert into collection
		if err = uc.usersRepository.GetUsersRepository().Insert(user); err != nil {
			panic(err)
		}
		// find user in collection
		if err = uc.usersRepository.FindUserByDeviceToken(request.DeviceToken, &user); err != nil {
			// if not found this time, there is an error
			panic(err)
		}
	}

	response := usersIO.CreateUserResponse{}
	response.User = user

	basecontroller.SendResponse(w, response)
}