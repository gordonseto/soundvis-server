package listeningsessionsrepository

import (
	"gopkg.in/mgo.v2"
	"sync"
	"github.com/gordonseto/soundvis-server/dbsession"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"github.com/gordonseto/soundvis-server/recordings/repositories/recordingsrepository"
	"github.com/gordonseto/soundvis-server/listeningsessions/models"
	"gopkg.in/mgo.v2/bson"
	"github.com/gordonseto/soundvis-server/config"
)

type ListeningSessionsRepository struct {
	session *mgo.Session
}

var instance *ListeningSessionsRepository
var once sync.Once

func Shared() *ListeningSessionsRepository {
	once.Do(func() {
		instance = &ListeningSessionsRepository{dbsession.Shared()}
	})
	return instance
}

func (rr *ListeningSessionsRepository) GetListeningSessionsRepository() *mgo.Collection {
	return rr.session.DB(config.DB_NAME).C("listening_sessions")
}

// streamPlayed is an id of a station or recording
func (lsr *ListeningSessionsRepository) InsertListeningSession(userId string, streamPlayed string, duration int64, startTime int64) error {
	// get the stationId from currentPlaying
	var stationId string
	if streamhelper.CurrentPlayingIsRecording(streamPlayed) {
		// if currentPlaying is a recording, get the recording
		recording, err := recordingsrepository.Shared().FindRecordingById(streamPlayed)
		if err != nil {
			return err
		}
		stationId = recording.StationId
	} else {
		stationId = streamPlayed
	}
	listeningSession := &listeningsession.ListeningSession{
		Id: bson.NewObjectId(),
		UserId: userId,
		StationId: stationId,
		Duration: duration,
		StartTime: startTime,
	}
	return Shared().GetListeningSessionsRepository().Insert(listeningSession)
}

func (lsr *ListeningSessionsRepository) FindSessionsByUser(userId string) ([]listeningsession.ListeningSession, error) {
	var ls []listeningsession.ListeningSession
	err := Shared().GetListeningSessionsRepository().Find(bson.M{"userId": userId}).All(&ls)
	return ls, err
}