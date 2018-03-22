package recordingscontroller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gordonseto/soundvis-server/authentication"
	"github.com/gordonseto/soundvis-server/general"
	"github.com/gordonseto/soundvis-server/jobmanager"
	"github.com/gordonseto/soundvis-server/recordings/IO"
	"github.com/gordonseto/soundvis-server/recordings/models"
	"github.com/gordonseto/soundvis-server/recordings/repositories/recordingsrepository"
	"github.com/gordonseto/soundvis-server/recordingsstream"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"sync"
	"time"
	"github.com/gordonseto/soundvis-server/stations/models"
	models2 "github.com/gordonseto/soundvis-server/users/models"
	"github.com/gordonseto/soundvis-server/users/repositories"
	"gopkg.in/mgo.v2/bson"
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

func (rc *RecordingsController) PUTPath() string {
	return "/recordings"
}

func (rc *RecordingsController) DELETEPath() string {
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

	var waitGroup sync.WaitGroup

	// for each recording, get the station corresponding to their stationId
	for _, recording := range recordings {
		waitGroup.Add(1)
		go GetStationForRecording(recording, &waitGroup)
	}

	// wait for all stations to be fetched for their recording
	waitGroup.Wait()

	response := recordingsIO.GetRecordingsResponse{}
	response.Recordings = recordings
	basecontroller.SendResponse(w, response)
}

func GetStationForRecording(recording *recordings.Recording, waitGroup *sync.WaitGroup) {
	station, err := streamhelper.GetStation(recording.StationId)
	if err != nil {
		panic(err)
	}
	recording.Station = station
	waitGroup.Done()
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

	if request.StationId == "" || request.EndDate == 0 {
		panic(errors.New("StationId or EndDate missing"))
	}

	// if start date has already passed, set as now
	now := time.Now().Unix()
	if request.StartDate < now {
		request.StartDate = now
	}

	if request.StartDate > request.EndDate || request.EndDate < now {
		panic(errors.New("EndDate must be in the future and EndDate must be after StartDate"))
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
	recording := &recordings.Recording{
		Id:        recordings.CreateRecordingId(),
		Title:     request.Title,
		CreatorId: user.Id.Hex(),
		StationId: request.StationId,
		StartDate: request.StartDate,
		EndDate:   request.EndDate,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		Status:    recordings.StatusPending,
	}

	// add recording job
	jobId, err := jobmanager.Shared().AddRecordingJob(recording)
	if err != nil {
		panic(err)
	}
	recording.JobId = jobId

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

func (rc *RecordingsController) UpdateRecording(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	request := recordingsIO.UpdateRecordingRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// make sure request has recordingId
	if request.RecordingId == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "RecordingId is required")
		return
	}
	recordingId := request.RecordingId

	// get recording
	recording, err := recordingsrepository.Shared().FindRecordingById(recordingId)
	if err != nil {
		panic(err)
	}

	// check user has permission
	if recording.CreatorId != user.Id.Hex() {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Permission denied")
		return
	}

	// update the recording title to match request
	recording.Title = request.Title
	recording.UpdatedAt = time.Now().Unix()

	var station *models.Station

	// if updating stationId, startDate and endDate, create new job and invalidate old one
	if request.StationId != "" || request.StartDate != 0 || request.EndDate != 0 {
		now := time.Now().Unix()
		// make sure recording has not already started
		if recording.StartDate < now || recording.EndDate < now || recording.Status != recordings.StatusPending {
			panic(errors.New("Cannot update recording that has already started"))
		}

		// if start date has already passed, set as now
		if request.StartDate < now {
			request.StartDate = now
		}

		// make sure request has valid start and end dates
		if request.StartDate > request.EndDate || request.EndDate < time.Now().Unix() {
			panic(errors.New("StartDate and EndDate must be in the future and EndDate must be after StartDate"))
		}

		if request.StationId == "" {
			panic(errors.New("StationId is required if updating startDate and endDate"))
		}
		// get station to make sure valid
		station, err = streamhelper.GetStation(request.StationId)
		if err != nil {
			panic(err)
		}

		// update the recording with new fields
		recording.StartDate = request.StartDate
		recording.EndDate = request.EndDate
		recording.StationId = request.StationId
		// add new recording job
		jobId, err := jobmanager.Shared().AddRecordingJob(recording)
		if err != nil {
			panic(err)
		}
		recording.JobId = jobId
	}

	// update recording in repository
	err = recordingsrepository.Shared().UpdateRecording(recording)
	if err != nil {
		panic(err)
	}

	// need to fill response with station
	if station == nil {
		station, err = streamhelper.GetStation(recording.StationId)
		if err != nil {
			panic(err)
		}
	}
	recording.Station = station

	// send response
	response := recordingsIO.CreateRecordingResponse{}
	response.Recording = recording
	basecontroller.SendResponse(w, response)
}

func (rc *RecordingsController) DeleteRecording(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user, err := authentication.CheckAuthentication(r)
	if err != nil {
		panic(err)
	}

	request := recordingsIO.DeleteRecordingRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// make sure request has recordingId
	if request.RecordingId == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "RecordingId is required")
		return
	}

	recordingId := request.RecordingId

	// get recording
	recording, err := recordingsrepository.Shared().FindRecordingById(recordingId)
	if err != nil {
		panic(err)
	}

	// check user has permission
	if recording.CreatorId != user.Id.Hex() {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Permission denied")
		return
	}

	// if recording file has already been created, delete
	if recording.Status == recordings.StatusFinished && recording.RecordingURL != "" {
		err = recordingsstream.DeleteRecordingFile(recordingId)
		if err != nil {
			panic(err)
		}
	}

	// delete recording from the repository
	err = recordingsrepository.Shared().GetRecordingsRepository().RemoveId(recordingId)
	if err != nil {
		panic(err)
	}

	// find all users currently playing recording and set their currentPlaying to empty
	var users []models2.User
	usersrepository.Shared().GetUsersRepository().Find(bson.M{"currentPlaying": recordingId}).All(&users)
	for _, user := range users {
		user.CurrentPlaying = ""
		user.IsPlaying = false
		usersrepository.Shared().UpdateUser(&user)
	}

	response := recordingsIO.DeleteRecordingResponse{}
	response.Ok = true

	basecontroller.SendResponse(w, response)
}