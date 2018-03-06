package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/gordonseto/soundvis-server/stations/controllers"
	"github.com/gordonseto/soundvis-server/config"
	"fmt"
	"encoding/json"
	"github.com/gordonseto/soundvis-server/users/controllers"
	"gopkg.in/mgo.v2"
)

func main() {
	r := httprouter.New()
	r.PanicHandler = handleError

	dbSession := getSession()

	sc := stations.NewStationsController()
	uc := users.NewUsersController(dbSession)

	r.GET(sc.GETPath(), sc.GetStations)
	r.POST(uc.POSTPath(), uc.CreateUser)

	http.ListenAndServe(config.PORT, r)
}

func getSession() *mgo.Session {
	s, err := mgo.Dial(config.DB_ADDRESS)

	if err != nil {
		panic(err)
	}
	return s
}

func handleError(w http.ResponseWriter, r *http.Request, err interface{}) {
	fmt.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	errorString := fmt.Sprintf("%s", err)

	type ErrorMessage struct {
		Error string	`json:"error"`
	}

	errorJSON, err := json.Marshal(&ErrorMessage{errorString})
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}

	w.Write([]byte(errorJSON))
}