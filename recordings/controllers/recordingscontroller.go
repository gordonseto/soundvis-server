package recordings

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/recordings/repositories"
	"github.com/gordonseto/soundvis-server/recordings/IO"
	"github.com/gordonseto/soundvis-server/general"
	"encoding/json"
	"errors"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"github.com/gordonseto/soundvis-server/recordings/models"
	"gopkg.in/mgo.v2/bson"
	"time"
	"github.com/gordonseto/soundvis-server/jobmanager"
)

type (
	RecordingsController struct {
	}
)

func (rc *RecordingsController) GETPath() string {
	return "/recordings"
}

func (rc *RecordingsController) POSTPath() string {
	return "/recordings"
}

func NewRecordingsController() *RecordingsController {
	return &RecordingsController{}
}

func (rc *RecordingsController) GetRecordings(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	recordings, err := recordingsrepository.Shared().FindRecordingsByCreatorId(user.Id.Hex())
	if err != nil {
		panic(err)
	}

	// for each recording, get the station corresponding to their stationId
	for _, recording := range recordings {
		recording.Station, err = streamhelper.GetStation(recording.StationId)
		if err != nil {
			panic(err)
		}
	}

	response := recordingsIO.GetRecordingsResponse{}
	response.Recordings = recordings
	basecontroller.SendResponse(w, response)
}

func (rc *RecordingsController) CreateRecording(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	// parse request
	request := recordingsIO.CreateRecordingRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	if request.StationId == "" || request.StartDate == 0 || request.EndDate == 0 {
		panic(errors.New("StationId, StartDate or EndDate missing"))
	}

	if request.StartDate > request.EndDate || request.StartDate < time.Now().Unix() || request.EndDate < time.Now().Unix() {
		panic(errors.New("StartDate and EndDate must be in the future and EndDate must be after StartDate"))
	}

	// get station to make sure valid
	station, err := streamhelper.GetStation(request.StationId)
	if err != nil {
		panic(err)
	}

	// if title is empty, take title from station name
	if request.Title == "" {
		request.Title = station.Name
	}
	recording := &models.Recording{
		Id: bson.NewObjectId().Hex(),
		Title: request.Title,
		CreatorId: user.Id.Hex(),
		StationId: request.StationId,
		StartDate: request.StartDate,
		EndDate: request.EndDate,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	// add recording job
	err = jobmanager.Shared().AddRecordingJob(recording)
	if err != nil {
		panic(err)
	}

	// insert into repository
	err = recordingsrepository.Shared().GetRecordingsRepository().Insert(recording)
	if err != nil {
		panic(err)
	}

	// send response
	response := recordingsIO.CreateRecordingResponse{}
	response.Recording = recording
	response.Recording.Station = station
	basecontroller.SendResponse(w, response)
}