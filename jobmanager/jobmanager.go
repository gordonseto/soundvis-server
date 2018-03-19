package jobmanager

import (
	"github.com/gordonseto/soundvis-server/streamjobs"
	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"os"
	"os/signal"
	"github.com/gordonseto/soundvis-server/config"
	"sync"
	"github.com/gordonseto/soundvis-server/recordingjobsmanager"
	"github.com/gordonseto/soundvis-server/recordings/models"
	"log"
	"time"
	"github.com/gordonseto/soundvis-server/recordings/repositories"
)

type JobManager struct {
}

type Context struct {
}

var enqueuer *work.Enqueuer
var redisPool *redis.Pool
var pool *work.WorkerPool

var STREAM_JOB_INTERVAL int64 = 5

var instance *JobManager
var once sync.Once

func Shared() *JobManager {
	once.Do(func() {
		// create the redisPool
		redisPool = &redis.Pool{
			MaxActive: 5,
			MaxIdle: 5,
			Wait: true,
			Dial: func() (redis.Conn, error ){
				return redis.Dial("tcp", ":6379")
			},
		}
		// create the enqueuer, this is to add tasks to
		enqueuer = work.NewEnqueuer(config.REDIS_NAMESPACE, redisPool)

		// this is the workerPool to executes tasks
		pool = work.NewWorkerPool(Context{}, 10, config.REDIS_NAMESPACE, redisPool)

		instance = &JobManager{}
	})
	return instance
}

// starts the jobManager
func (jm *JobManager) Start() {
	pool.Start()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	pool.Stop()
}

// creates streamJobs that will run every STREAM_JOB_INTERVAL
func (jm *JobManager) RegisterStreamJobs() {
	// add handler for job
	pool.Job(streamjobmanager.StreamJobName(), (*Context).runStreamJob)

	// enqueue initial stream job
	jm.enqueueStreamJob()
}

func (c *Context) runStreamJob(job *work.Job) error {
	// enqueue a new streamJob to run after this one is finished
	Shared().enqueueStreamJob()
	// check the now playing of all current users
	streamjobmanager.Shared().RefreshNowPlaying()
	return nil
}

// enqueues a new streamJob to run
func (jm *JobManager) enqueueStreamJob() error {
	_, err := enqueuer.EnqueueUniqueIn(streamjobmanager.StreamJobName(), STREAM_JOB_INTERVAL, nil)
	return err
}

// adds recordingJob handler to jobManager
func (jm *JobManager) RegisterRecordingJobs() {
	pool.Job(recordingjobsmanager.RecordingJobName(), (*Context).runRecordingJob)
}

// adds a new recordingjob to be executed at recording.StartDate
func (jm *JobManager) AddRecordingJob(recording *models.Recording) error {
	log.Println("Adding recordingJob - recordingId: " + recording.Id + ", set to run at ", recording.StartDate)

	now := time.Now().Unix()
	secondsFromNow := recording.StartDate - now
	if secondsFromNow < 0 {
		secondsFromNow = 0
	}

	// enqueue the job
	_, err := enqueuer.EnqueueIn(recordingjobsmanager.RecordingJobName(), secondsFromNow, work.Q{"id": recording.Id, "stationId": recording.StationId, "startDate": recording.StartDate, "endDate": recording.EndDate})
	return err
}

// runs the recording job
func (c *Context) runRecordingJob(job *work.Job) error {
	// get parameters saved when job was enqueued
	id := job.ArgString("id")
	stationId := job.ArgString("stationId")
	startDate := job.ArgInt64("startDate")
	endDate := job.ArgInt64("endDate")
	log.Println("Running recordingJob - recordingId: " + id)

	if endDate < time.Now().Unix() {
		log.Println("Recording end date has already passed, finishing - recordingId: " + id)
		return nil
	}

	// record the stream
	err := recordingjobsmanager.Shared().RecordStream(id, stationId, startDate, endDate)
	if err != nil {
		log.Println("Error for recording job, recordingId: ", id)
		log.Println(err)
		err = recordingsrepository.Shared().UpdateRecordingStatus(id, models.StatusFailed)
		if err != nil {
			log.Println(err)
		}
	}
	return err
}