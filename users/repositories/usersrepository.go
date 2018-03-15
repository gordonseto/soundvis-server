package usersrepository

import (
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/config"
	"gopkg.in/mgo.v2/bson"
	"github.com/gordonseto/soundvis-server/users/models"
	"errors"
)

type UsersRepository struct {
	session *mgo.Session
}

func NewUsersRepository(s *mgo.Session) *UsersRepository {
	return &UsersRepository{s}
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