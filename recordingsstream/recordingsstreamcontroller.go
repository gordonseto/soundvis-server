package recordingsstream

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"bytes"
)

type RecordingsStreamController struct {
}

func NewRecordingsStreamController() *RecordingsStreamController {
	return &RecordingsStreamController{}
}

func (rsc *RecordingsStreamController) GETPath() string {
	return "/recordings/stream/:recordingId"
}

func (rsc *RecordingsStreamController) StreamRecording(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// recordingId is in URL parameter
	recordingId := p.ByName("recordingId")
	// get the fileName
	fileName := GetRecordingFileNameFromId(recordingId)

	// read the file
	recordingBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	b := bytes.NewBuffer(recordingBytes)
	// set the header
	w.Header().Set("Content-Type", "audio/mpeg")

	// stream the bytes to client
	if _, err := b.WriteTo(w); err != nil {
		panic(err)
	}
	return
}

func GetFilePath() string {
	return "files/recordings/"
}

func GetRecordingFileNameFromId(recordingId string) string {
	return GetFilePath() + recordingId + ".mp3"
}

func GetRecordingStreamPath(recordingId string) string {
	return "/recordings/stream/" + recordingId
}