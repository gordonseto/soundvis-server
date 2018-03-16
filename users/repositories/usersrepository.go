package usersrepository

import (
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/config"
	"gopkg.in/mgo.v2/bson"
	"github.com/gordonseto/soundvis-server/users/models"
	"errors"
	"sync"
	"github.com/gordonseto/soundvis-server/dbsession"
)

type UsersRepository struct {
	session *mgo.Session
}

var instance *UsersRepository
var once sync.Once

func Shared() *UsersRepository {
	once.Do(func() {
		instance = &UsersRepository{dbsession.Shared()}
	})
	return instance
}

func (ur *UsersRepository) GetUsersRepository() *mgo.Collection {
	return ur.session.DB(config.DB_NAME).C("stations")
}

func (ur *UsersRepository) FindUserById(userId string) (*models.User, error) {
	var user models.User
	if !bson.IsObjectIdHex(userId) {
		return nil, errors.New("Invalid userId")
	}

	oid := bson.ObjectIdHex(userId)

	err := ur.GetUsersRepository().FindId(oid).One(&user)
	return &user, err
}

func (ur *UsersRepository) FindUserByDeviceToken(deviceToken string, user *models.User) error {
	return ur.GetUsersRepository().Find(bson.M{"deviceToken":deviceToken}).One(&user)
}


func (ur *UsersRepository) UpdateUser(user *models.User) error {
	return ur.GetUsersRepository().Update(bson.M{"_id": user.Id}, user)
}

func (ur *UsersRepository) FindUsersByIsPlaying(isPlaying bool) ([]models.User, error) {
	users := make([]models.User, 0)
	err := ur.GetUsersRepository().Find(bson.M{"isPlaying": isPlaying}).All(&users)
	return users, err
}