package recordingjobsmanager

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
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

func (rjm *RecordingJobsManager) RecordStream() {
	streamURL := "http://138.201.248.219:8018/stream"
	startDate := time.Now().Unix()
	endDate := time.Now().Unix() + 30

	file, err := os.Create("output.mp3")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	for time.Now().Unix() < startDate {
		fmt.Println("waiting for startDate")
	}

	resp, err := http.Get(streamURL)
	if err != nil {
		fmt.Println(err)
	}
	reader := bufio.NewReader(resp.Body)
	for time.Now().Unix() < endDate {
		b, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
		} else {
			log.Println(b)
			if _, err := file.Write([]byte{b}); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (rjm *RecordingJobsManager) Test() {
	fmt.Println("Hello World!")

	file, err := os.Create("output.mp3")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	start := time.Now().Unix()

	resp, err := http.Get("http://138.201.248.219:8018/stream")
	if err != nil {
		fmt.Println(err)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
		} else {
			log.Println(b)
			if _, err := file.Write([]byte{b}); err != nil {
				fmt.Println(err)
			}
		}
		now := time.Now().Unix()
		if now-start > 10 {
			return
		}
	}
}
