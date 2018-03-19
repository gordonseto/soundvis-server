package recordingtracklistsrepository

import (
	"gopkg.in/mgo.v2"
	"sync"
	"github.com/gordonseto/soundvis-server/dbsession"
	"github.com/gordonseto/soundvis-server/config"
	"github.com/gordonseto/soundvis-server/recordings/models"
	"gopkg.in/mgo.v2/bson"
)

type RecordingTracklistsRepository struct {
	session *mgo.Session
}

var instance *RecordingTracklistsRepository
var once sync.Once

func Shared() *RecordingTracklistsRepository {
	once.Do(func() {
		instance = &RecordingTracklistsRepository{dbsession.Shared()}
	})
	return instance
}

func (rtr *RecordingTracklistsRepository) GetRecordingTracklistsRepository() *mgo.Collection {
	return rtr.session.DB(config.DB_NAME).C("recording_tracklists")
}

func (rtr *RecordingTracklistsRepository) FindTracklistByRecordingId(recordingId string) (*recordings.RecordingTrackList, error) {
	var tracklist recordings.RecordingTrackList
	err := rtr.GetRecordingTracklistsRepository().Find(bson.M{"recordingId": recordingId}).One(&tracklist)
	return &tracklist, err
}
