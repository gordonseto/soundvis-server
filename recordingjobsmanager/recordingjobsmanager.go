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
	"gopkg.in/mgo.v2/bson"
	"github.com/gordonseto/soundvis-server/recordings/models"
	"github.com/gordonseto/soundvis-server/stationsfetcher"
	"github.com/gordonseto/soundvis-server/stream/models"
	"github.com/gordonseto/soundvis-server/recordings/repositories/recordingsrepository"
	"github.com/gordonseto/soundvis-server/recordings/repositories/recordingtracklistsrepository"
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

	// wait if startDate has not passed
	for time.Now().Unix() < startDate {
		fmt.Println("Waiting for startDate, recordingId: " + recordingId)
	}

	log.Println("Duration: ", endDate - time.Now().Unix())

	// update recording status to IN_PROGRESS
	err = recordingsrepository.Shared().UpdateRecordingStatus(recordingId, recordings.StatusInProgress)
	if err != nil {
		return err
	}

	// create channels for the 2 async tasks
	streamChannel := make(chan []byte, 1)
	tracklistChannel := make(chan *recordings.RecordingTrackList, 1)

	// spin a thread to stream the audio bytes
	go func() {
		stream, e := rjm.recordAudio(streamURL, endDate)
		err = e
		streamChannel <- stream
	}()

	// spin a thread to record the tracklist
	go func() {
		tracklist := rjm.recordTracklist(recordingId, streamURL, startDate, endDate)
		tracklistChannel <- tracklist
	}()

	// wait for stream and tracklist to finish recording
	stream := <-streamChannel
	tracklist := <-tracklistChannel
	if err != nil {
		return err
	}

	log.Println("Done recording for recordingId: ", recordingId)

	// save recording to file
	err = rjm.saveStreamToFile(recordingId, stream)
	if err != nil {
		panic(err)
	}

	// update recording with recordingURL and status
	recordingURL := recordingsstream.GetRecordingStreamPath(recordingId)
	err = recordingsrepository.Shared().GetRecordingsRepository().UpdateId(recordingId, bson.M{"$set": bson.M{"recordingUrl": recordingURL, "status": recordings.StatusFinished}})
	if err != nil {
		log.Println(err)
		// if failure, delete the recording
		err = recordingsstream.DeleteRecordingFile(recordingId)
		return err
	}

	// insert recording tracklist into DB
	err = recordingtracklistsrepository.Shared().GetRecordingTracklistsRepository().Insert(tracklist)
	if err != nil {
		return err
	}

	log.Println("Finished processing recordingId: ", recordingId)
	return nil
}

// records the audio from reader from now until endDate
func (rjm *RecordingJobsManager) recordAudio(streamURL string, endDate int64) ([]byte, error) {
	// begin streaming from station
	resp, err := http.Get(streamURL)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(resp.Body)

	stream := make([]byte, 0)
	// keep streaming until endDate has passed
	for time.Now().Unix() < endDate {
		// read the audio byte
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		} else {
			if time.Now().Unix() > endDate {
				break
			}
			// append to array
			stream = append(stream, b)
		}
	}
	return stream, nil
}

// records the tracklist of a streamURL from now until endDate
func (rjm *RecordingJobsManager) recordTracklist(recordingId string, streamURL string, startDate int64, endDate int64) *recordings.RecordingTrackList {
	tracklist := recordings.NewRecordingTrackList(recordingId)
	SLEEP_DURATION := 3 * time.Second
	var lastSong *stream.Song

	// loop duration of recording
	for time.Now().Unix() < endDate {
		// get the current song
		song, err := stationsfetcher.GetCurrentSongPlayingShoutcast(streamURL)
		if err != nil {
			log.Println(err)
		}
		// compare if same as lastSong
		if lastSong == nil || (song.Title != lastSong.Title && song.Name != lastSong.Name) {
			// if song has changed, add to tracklist
			progress := time.Now().Unix() - startDate
			tracklist.AddTimeStamp(progress, song)
		}
		lastSong = song
		// sleep
		time.Sleep(SLEEP_DURATION)
	}
	return tracklist
}

func (rjm *RecordingJobsManager) saveStreamToFile(recordingId string, stream []byte) error {
	// if folder does not exist, create
	if _, err := os.Stat(recordingsstream.GetFilePath()); os.IsNotExist(err) {
		os.MkdirAll(recordingsstream.GetFilePath(), os.ModePerm)
	}
	// create the file
	fileName := recordingsstream.GetRecordingFileNameFromId(recordingId)
	file, err := os.Create(fileName)
	if err != nil {
		return errors.New("Error creating file for recordingId: " + recordingId)
	}
	defer file.Close()

	// save byte array
	if _, err := file.Write(stream); err != nil {
		return err
	}
	return nil
}