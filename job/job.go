package job

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/egorka-gh/photocycle"
	log "github.com/go-kit/kit/log"
)

//NewRuner creates Runer
func NewRuner(interval int, repo photocycle.Repository, logger log.Logger, jobs ...Job) Runer {
	if interval < 3 {
		interval = 3
	}
	r := baseRuner{
		interval: interval,
		repo:     repo,
		logger:   logger,
		jobs:     jobs,
	}
	return &r
}

//Job job to do
type Job interface {
	Init() error
	Do(ctx context.Context)
}

//FillBox creates FillBox job
func FillBox() Job {
	return &baseJob{
		name:   "FillBox",
		doFunc: fillBoxes,
	}
}

type baseJob struct {
	name     string
	repo     photocycle.Repository
	logger   log.Logger
	initFunc func(j *baseJob) error
	doFunc   func(ctx context.Context, j *baseJob) error
}

func (j *baseJob) Init() error {
	if j.logger == nil {
		j.logger = log.NewNopLogger()
	}
	j.logger = log.With(j.logger, "job", j.name)
	if j.initFunc != nil {
		return j.initFunc(j)
	}
	return nil
}

func (j *baseJob) Do(ctx context.Context) {
	if j.doFunc != nil {
		err := j.doFunc(ctx, j)
		if err != nil {
			j.logger.Log("Error", err)
		}
	}
	return
}

//Runer job runer
type Runer interface {
	Run(quit chan struct{}) error
}

//Runer job runer implementation
type baseRuner struct {
	interval int
	repo     photocycle.Repository
	logger   log.Logger
	jobs     []Job
}

//Run runs jobs periodicaly, blocks caller till get quit
func (r *baseRuner) Run(quit chan struct{}) error {
	if len(r.jobs) == 0 {
		err := errors.New("No jobs to do")
		return err
	}
	if r.logger == nil {
		r.logger = log.NewNopLogger()
	}
	r.logger = log.With(r.logger, "actor", "runner")
	r.logger.Log("event", "Starting.")
	//init jobs
	r.logger.Log("event", "Init jobs.")
	for _, job := range r.jobs {
		if j, ok := job.(*baseJob); ok {
			j.repo = r.repo
			j.logger = r.logger
		}
		if err := job.Init(); err != nil {
			return err
		}
	}

	var timer *time.Timer
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()
	start := make(chan int, 1)
	defer close(start)
	var wg sync.WaitGroup
	loop := true

	start <- 1
	for loop {
		select {
		case <-start:
			ctx, cancel := context.WithCancel(mainCtx)
			wg.Add(1)
			r.logger.Log("event", "Starting jobs.")
			go func() {
				defer wg.Done()
				defer cancel()
				//run sequentially
				for _, job := range r.jobs {
					job.Do(ctx)
					if ctx.Err() != nil {
						//contex canceled
						return
					}
				}
				timer = time.AfterFunc(time.Minute*time.Duration(r.interval), func() { start <- 1 })
			}()
		case <-quit:
			if timer != nil {
				timer.Stop()
			}
			mainCancel()
			loop = false
			r.logger.Log("event", "Stop")
			wg.Wait()
		}
	}

	return nil
}
