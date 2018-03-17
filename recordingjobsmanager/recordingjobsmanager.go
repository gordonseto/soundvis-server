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

func (rjm *RecordingJobsManager) RecordStream(recordingId string, stationId string, startDate int64, endDate int64) error {
	station, err := streamhelper.GetStation(stationId)
	if err != nil {
		return errors.New("Error getting station in recording job, recordingId: " + recordingId)
	}
	streamURL := station.StreamURL

	file, err := os.Create("output.mp3")
	if err != nil {
		return errors.New("Error creating file for recordingId: " + recordingId)
	}
	defer file.Close()

	for time.Now().Unix() < startDate {
		fmt.Println("Waiting for startDate, recordingId: " + recordingId)
	}

	log.Println("Duration: ", endDate - time.Now().Unix())

	resp, err := http.Get(streamURL)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(resp.Body)
	for time.Now().Unix() < endDate {
		b, err := reader.ReadByte()
		if err != nil {
			return err
		} else {
			if time.Now().Unix() > endDate {
				break
			}
			if _, err := file.Write([]byte{b}); err != nil {
				return err
			}
		}
	}
	log.Println("Done recording for recordingId: ", recordingId)
	return nil
}