package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/gordonseto/soundvis-server/stations/controllers"
	"github.com/gordonseto/soundvis-server/config"
	"fmt"
	"encoding/json"
	"github.com/gordonseto/soundvis-server/users/controllers"
	"github.com/gordonseto/soundvis-server/stream/controllers"
	//"github.com/gordonseto/soundvis-server/stationsfetcher"
	"github.com/gordonseto/soundvis-server/socketmanager"
	"github.com/gordonseto/soundvis-server/jobmanager"
	"github.com/gordonseto/soundvis-server/recordingsstream"
	"github.com/gordonseto/soundvis-server/recordings/controllers"
	"github.com/gordonseto/soundvis-server/listeningsessions/controllers"
)

func main() {
	r := httprouter.New()
	r.PanicHandler = handleError

	stationsController := stations.NewStationsController()
	usersController := users.NewUsersController()
	streamsController := stream.NewStreamController()
	recordingsController := recordingscontroller.NewRecordingsController()
	recordingsStreamController := recordingsstream.NewRecordingsStreamController()
	listeningSessionsController := listeningsessionscontroller.NewListeningSessionsController()

	r.GET(stationsController.GETPath(), stationsController.GetStations)
	r.POST(usersController.POSTPath(), usersController.CreateUser)
	r.GET(streamsController.GETPath(), streamsController.GetCurrentStream)
	r.POST(streamsController.POSTPath(), streamsController.SetCurrentStream)
	r.GET(recordingsController.GETPath(), recordingsController.GetRecordings)
	r.POST(recordingsController.POSTPath(), recordingsController.CreateRecording)
	r.PUT(recordingsController.PUTPath(), recordingsController.UpdateRecording)
	r.DELETE(recordingsController.DELETEPath(), recordingsController.DeleteRecording)
	r.GET(recordingsStreamController.GETPath(), recordingsStreamController.StreamRecording)
	r.GET(listeningSessionsController.GETPath(), listeningSessionsController.GetStats)

	r.POST(socketmanager.Shared().POSTPath(), socketmanager.Shared().Connect)
	r.GET(socketmanager.Shared().GETPath(), socketmanager.Shared().Connect)
	//stationsfetcher.FetchAndStoreStations(stationsrepository.Shared())

	jobmanager.Shared().RegisterStreamJobs()
	jobmanager.Shared().RegisterRecordingJobs()

	go func(){
		jobmanager.Shared().Start()
	}()

	http.ListenAndServe(config.PORT, r)
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