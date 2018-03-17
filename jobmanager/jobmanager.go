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

func (jm *JobManager) RegisterRecordingJobs() {
	pool.Job(recordingjobsmanager.RecordingJobName(), (*Context).runRecordingJob)
}

func (jm *JobManager) AddRecordingJob(recording *models.Recording) error {
	log.Println("Adding recordingJob - recordingId: " + recording.Id.Hex() + ", set to run at ", recording.StartDate)

	now := time.Now().Unix()
	secondsFromNow := recording.StartDate - now
	if secondsFromNow < 0 {
		secondsFromNow = 0
	}

	_, err := enqueuer.EnqueueIn(recordingjobsmanager.RecordingJobName(), secondsFromNow, work.Q{"id": recording.Id.Hex(), "stationId": recording.StationId, "startDate": recording.StartDate, "endDate": recording.EndDate})
	return err
}

func (c *Context) runRecordingJob(job *work.Job) error {
	id := job.ArgString("id")
	stationId := job.ArgString("stationId")
	startDate := job.ArgInt64("startDate")
	endDate := job.ArgInt64("endDate")
	log.Println("Running recordingJob - recordingId: " + id)

	err := recordingjobsmanager.Shared().RecordStream(id, stationId, startDate, endDate)
	if err != nil {
		log.Println("Error for recording job, recordingId: ", id)
		log.Println(err)
	}
	return err
}