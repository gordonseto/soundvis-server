package stationsrepository

import (
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/config"
)

type StationsRepository struct {
	session *mgo.Session
}

func NewStationsRepository(s *mgo.Session) *StationsRepository {
	return &StationsRepository{s}
}

func (sr *StationsRepository) GetStationsRepository() *mgo.Collection {
	return sr.session.DB(config.DB_NAME).C("stations")
}