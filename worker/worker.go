package worker

import (
	"errors"
	"github.com/oklahomer/go-sarah/log"
	"golang.org/x/net/context"
	"runtime"
	"sync"
)

/*
Worker holds desired number of child workers when Run is called.
*/
type Worker struct {
	jobReceiver chan func()
	mutex       *sync.Mutex
	isRunning   bool
}

/*
New is a helper function that construct and return new Worker instance.
*/
func New() *Worker {
	return &Worker{
		jobReceiver: make(chan func(), 100),
		mutex:       &sync.Mutex{},
		isRunning:   false,
	}
}

/*
Run creates as many child workers as specified and start those child workers.
First argument, cancel channel, can be context.Context.Done to propagate upstream status change.
*/
func (worker *Worker) Run(ctx context.Context, workerNum uint) error {
	log.Infof("start workers")
	worker.mutex.Lock()
	defer worker.mutex.Unlock()

	if worker.isRunning == true {
		return errors.New("workers are already running")
	}

	var i uint
	for i = 1; i <= workerNum; i++ {
		go worker.runChild(ctx, i)
	}
	worker.isRunning = true

	// update status to false on cancellation
	go func() {
		<-ctx.Done()
		worker.isRunning = false
	}()

	return nil
}

func (worker *Worker) runChild(ctx context.Context, workerId uint) {
	log.Infof("start worker id: %d.", workerId)

	for {
		select {
		case <-ctx.Done():
			log.Infof("stopping worker id: %d", workerId)
			return
		case job := <-worker.jobReceiver:
			log.Debugf("receiving job on worker: %d", workerId)
			// To avoid given job's panic affect later jobs, wrap them with recover.
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Warnf("panic in given job. recovered: %#v", r)

						// Display stack trace
						for depth := 0; ; depth++ {
							_, src, line, ok := runtime.Caller(depth)
							if !ok {
								break
							}
							log.Warnf(" -> depth:%d. file:%s. line:%d.", depth, src, line)
						}
					}

				}()
				job()
			}()
		}
	}
}

/*
IsRunning returns current status of worker.
*/
func (worker *Worker) IsRunning() bool {
	return worker.isRunning
}

/*
EnqueueJob appends new job to be executed.
*/
func (worker *Worker) EnqueueJob(job func()) {
	worker.jobReceiver <- job
}
