package recordingsrepository

import (
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/config"
	"sync"
	"github.com/gordonseto/soundvis-server/dbsession"
	"github.com/gordonseto/soundvis-server/recordings/models"
	"gopkg.in/mgo.v2/bson"
)

type RecordingsRepository struct {
	session *mgo.Session
}

var instance *RecordingsRepository
var once sync.Once

func Shared() *RecordingsRepository {
	once.Do(func() {
		instance = &RecordingsRepository{dbsession.Shared()}
	})
	return instance
}

func (rr *RecordingsRepository) GetRecordingsRepository() *mgo.Collection {
	return rr.session.DB(config.DB_NAME).C("recordings")
}

func (rr *RecordingsRepository) FindRecordingById(recordingId string) (*models.Recording, error) {
	var recording *models.Recording
	err := rr.GetRecordingsRepository().Find(bson.M{"_id": recordingId}).One(&recording)
	return recording, err
}

func (rr *RecordingsRepository) FindRecordingsByCreatorId(creatorId string) ([]*models.Recording, error) {
	recordings := make([]*models.Recording, 0)
	err := rr.GetRecordingsRepository().Find(bson.M{"creatorId": creatorId}).All(&recordings)
	return recordings, err
}