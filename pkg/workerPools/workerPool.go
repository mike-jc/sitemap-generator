package workerPools

import (
	"sitemap-generator/services"
	"sync"
)

// WorkerHandler is a job, processor of the task
type WorkerHandler func(v interface{}) error

// WorkerPool processes tasks in parallel by limited number of the workers
type WorkerPool interface {
	Init(handler WorkerHandler) (startedWorkers int, err error)
	AddTask(v interface{})
	WaitFinalize()
}

type workerPool struct {
	workersCount int

	logger    services.Logger
	tasksChan chan interface{}
	jobs      sync.WaitGroup
	workers   sync.WaitGroup
}

func NewWorkerPool(logger services.Logger, workersCount int) WorkerPool {
	return &workerPool{
		logger:       logger,
		workersCount: workersCount,
	}
}

// Init starts specified number of workers which expect new tasks from the queue
func (wp *workerPool) Init(handler WorkerHandler) (startedWorkers int, err error) {
	wp.tasksChan = make(chan interface{}, wp.workersCount)

	runJob := func(task interface{}) {
		defer wp.jobs.Done()
		if err := handler(task); err != nil {
			wp.logger.Error("WorkerPool: could not succeed the job in the worker", err.Error())
		}
	}

	// start workers
	for i := 0; i < wp.workersCount; i++ {
		wp.workers.Add(1)
		startedWorkers++

		go func() {
			defer func() {
				wp.workers.Done()

				// todo: improve to restart worker somehow when runtime error to keep processing other tasks
				if r := recover(); r != nil {
					wp.logger.Error("WorkerPool: error in worker", r)
				}
			}()

			// read tasks from the queue while it's open
			for task := range wp.tasksChan {
				runJob(task)
			}
		}()
	}
	return startedWorkers, nil
}

func (wp *workerPool) AddTask(v interface{}) {
	wp.jobs.Add(1)
	wp.tasksChan <- v
}

// WaitFinalize waits until all tasks are processed and workers stopped
// and close the input/output channels
func (wp *workerPool) WaitFinalize() {
	wp.jobs.Wait()
	close(wp.tasksChan)
	wp.workers.Wait()
}
