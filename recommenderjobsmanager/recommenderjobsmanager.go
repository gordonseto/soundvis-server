package recommenderjobsmanager

import (
	"sync"
	"github.com/gordonseto/soundvis-server/listeningsessions/repositories"
	"gopkg.in/mgo.v2/bson"
	"log"
	"github.com/gordonseto/soundvis-server/scripthelper"
)

type RecommenderJobsManager struct {
	numSessions int
}


var instance *RecommenderJobsManager
var once sync.Once

func Shared() *RecommenderJobsManager {
	once.Do(func(){
		instance = &RecommenderJobsManager{}
	})
	return instance
}

func RecommenderJobName() string {
	return "recommender_job"
}

func (rjm *RecommenderJobsManager) RunTrainer() {
	numSessions, err := listeningsessionsrepository.Shared().GetListeningSessionsRepository().Find(bson.M{}).Count()
	if err != nil {
		log.Println(err)
		return
	}
	// if number of listening sessions has changed, re-run trainer
	if rjm.numSessions != numSessions {
		rjm.numSessions = numSessions
		log.Println("Running trainer...")
		_, err := scripthelper.RunTrainerScript()
		if err != nil {
			log.Println("Trainer script failed, error: ")
			log.Println(err)
			return
		}
		log.Println("Done running trainer")
	} else {
		rjm.numSessions = numSessions
	}
}