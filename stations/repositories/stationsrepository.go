package stationsrepository

import (
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/config"
	"github.com/gordonseto/soundvis-server/stations/models"
	"gopkg.in/mgo.v2/bson"
	"errors"
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

func (sr *StationsRepository) FindStationById(stationId string) (*models.Station, error) {
	var station models.Station
	if !bson.IsObjectIdHex(stationId) {
		return nil, errors.New("Invalid id")
	}

	oid := bson.ObjectIdHex(stationId)

	err := sr.GetStationsRepository().FindId(oid).One(&station)
	return &station, err
}