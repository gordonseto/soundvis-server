package dbsession

import (
	"gopkg.in/mgo.v2"
	"sync"
	"github.com/gordonseto/soundvis-server/config"
)

var instance *mgo.Session
var once sync.Once

func Shared() *mgo.Session {
	once.Do(func(){
		s, err := mgo.Dial(config.DB_ADDRESS)
		if err != nil {
			panic(err)
		}
		instance = s
	})
	return instance
}