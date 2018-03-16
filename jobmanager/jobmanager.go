package jobmanager

import (
	"github.com/gordonseto/soundvis-server/streamjobs"
	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"os"
	"os/signal"
	"github.com/gordonseto/soundvis-server/config"
)

type JobManager struct {
}

type Context struct {
	jobManager *JobManager
	streamJobManager *streamjobs.StreamJobManager
}

var enqueuer *work.Enqueuer
var redisPool *redis.Pool
var pool *work.WorkerPool

var STREAM_JOB_INTERVAL int64 = 5

func NewJobManager() *JobManager {
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

	jm := &JobManager{}

	// set context to always have jobManager
	pool.Middleware(func(c *Context, job *work.Job, next work.NextMiddlewareFunc) error {
		c.jobManager = jm
		return next()
	})

	return jm
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
func (jm *JobManager) RegisterStreamJobs(sjm *streamjobs.StreamJobManager) {
	// set context to always have streamJobManager
	pool.Middleware(func(c *Context, job *work.Job, next work.NextMiddlewareFunc) error {
		c.streamJobManager = sjm
		return next()
	})

	// add handler for job
	pool.Job(streamjobs.StreamJobName(), (*Context).runStreamJob)

	// enqueue initial stream job
	jm.enqueueStreamJob()
}

func (c *Context) runStreamJob(job *work.Job) error {
	// enqueue a new streamJob to run after this one is finished
	c.jobManager.enqueueStreamJob()
	c.streamJobManager.RefreshNowPlaying()
	return nil
}

// enqueues a new streamJob to run
func (jm *JobManager) enqueueStreamJob() error {
	_, err := enqueuer.EnqueueUniqueIn(streamjobs.StreamJobName(), STREAM_JOB_INTERVAL, nil)
	return err
}