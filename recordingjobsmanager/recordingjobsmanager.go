package recordingjobsmanager

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"github.com/gordonseto/soundvis-server/streamhelper"
	"errors"
	"github.com/gordonseto/soundvis-server/recordingsstream"
)

type RecordingJobsManager struct {
}

var instance *RecordingJobsManager
var once sync.Once
var mutex = &sync.Mutex{}

func Shared() *RecordingJobsManager {
	once.Do(func() {
		instance = &RecordingJobsManager{}
	})
	return instance
}

func RecordingJobName() string {
	return "recording_job"
}

// records audio from the station corresponding to stationId between startDate and endDate and saves to disk
func (rjm *RecordingJobsManager) RecordStream(recordingId string, stationId string, startDate int64, endDate int64) error {
	// get the station stationId belongs to
	station, err := streamhelper.GetStation(stationId)
	if err != nil {
		return errors.New("Error getting station in recording job, recordingId: " + recordingId)
	}
	streamURL := station.StreamURL


	// create the file
	fileName := recordingsstream.GetRecordingFileNameFromId(recordingId)
	file, err := os.Create(fileName)
	if err != nil {
		return errors.New("Error creating file for recordingId: " + recordingId)
	}
	defer file.Close()

	// wait if startDate has not passed
	for time.Now().Unix() < startDate {
		fmt.Println("Waiting for startDate, recordingId: " + recordingId)
	}

	log.Println("Duration: ", endDate - time.Now().Unix())

	// begin streaming from station
	resp, err := http.Get(streamURL)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(resp.Body)

	// keep streaming until endDate has passed
	for time.Now().Unix() < endDate {
		// read the audio byte
		b, err := reader.ReadByte()
		if err != nil {
			return err
		} else {
			if time.Now().Unix() > endDate {
				break
			}
			// save to file
			if _, err := file.Write([]byte{b}); err != nil {
				return err
			}
		}
	}
	log.Println("Done recording for recordingId: ", recordingId)
	return nil
}