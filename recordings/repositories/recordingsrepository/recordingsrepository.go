package recordingsrepository

import (
	"gopkg.in/mgo.v2"
	"github.com/gordonseto/soundvis-server/config"
	"sync"
	"github.com/gordonseto/soundvis-server/dbsession"
	"gopkg.in/mgo.v2/bson"
	"github.com/gordonseto/soundvis-server/recordings/models"
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

func (rr *RecordingsRepository) FindRecordingById(recordingId string) (*recordings.Recording, error) {
	var recording *recordings.Recording
	err := rr.GetRecordingsRepository().Find(bson.M{"_id": recordingId}).One(&recording)
	return recording, err
}

func (rr *RecordingsRepository) FindRecordingsByCreatorId(creatorId string) ([]*recordings.Recording, error) {
	recordings := make([]*recordings.Recording, 0)
	err := rr.GetRecordingsRepository().Find(bson.M{"creatorId": creatorId}).All(&recordings)
	return recordings, err
}

func (rr *RecordingsRepository) UpdateRecordingStatus(recordingId string, status string) error {
	return rr.GetRecordingsRepository().UpdateId(recordingId, bson.M{"$set": bson.M{"status": status}})
}