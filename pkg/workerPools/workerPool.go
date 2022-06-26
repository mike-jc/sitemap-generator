package workerPools

import (
	"fmt"
	"sitemap-generator/services"
	"sync"
)

type Identifiable interface {
	Id() string
}

// WorkerHandler is a job, processor of the task
type WorkerHandler func(v interface{}) error

// NeedsTaskFunc returns *true* if input requires a new task for processing
type NeedsTaskFunc func(v interface{}) bool

// WorkerPool allows to process tasks in parallel by limited number of the workers
type WorkerPool interface {
	Init(handler WorkerHandler) error
	Dispatch(v interface{}, needsTask NeedsTaskFunc)
	WaitFinalize()
	Results() ([]interface{}, error)
}

type workerPool struct {
	logger services.Logger

	workersCount int
	isFinalized  bool

	tasksChan chan interface{}
	jobs      sync.WaitGroup
	workers   sync.WaitGroup

	resultsLocker sync.Mutex
	results       map[string]Identifiable
}

func NewWorkerPool(logger services.Logger, workersCount int) WorkerPool {
	return &workerPool{
		logger:       logger,
		workersCount: workersCount,
	}
}

// Init starts specified number of workers which expect new tasks from the queue
func (wp *workerPool) Init(handler WorkerHandler) error {
	wp.tasksChan = make(chan interface{}, wp.workersCount)
	wp.results = make(map[string]Identifiable)

	runJob := func(task interface{}) {
		defer wp.jobs.Done()
		if err := handler(task); err != nil {
			wp.logger.Error("WorkerPool: could not succeed the job in the worker", err.Error())
		}
	}

	// start workers
	for i := 0; i < wp.workersCount; i++ {
		wp.workers.Add(1)

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
	return nil
}

// Dispatch adds the given value to the results if it's not there yet and
// adds a task to the queue for it if needed
func (wp *workerPool) Dispatch(v interface{}, needsTask NeedsTaskFunc) {
	if idn, ok := v.(Identifiable); ok {
		if wp.addResult(idn) && needsTask(idn) {
			// result has a task, create a job
			wp.jobs.Add(1)
			wp.tasksChan <- idn
		}
	} else {
		wp.logger.Warn(fmt.Sprintf("WorkerPool: wrong type of input portion for dispatching: %T", v))
	}
}

// addResult collects result if it's not yet in the list and returns *true*, or *false* otherwise;
// results list is being locked while reading from and writing in
func (wp *workerPool) addResult(idn Identifiable) bool {
	wp.resultsLocker.Lock()
	defer wp.resultsLocker.Unlock()

	if _, exists := wp.results[idn.Id()]; !exists {
		wp.results[idn.Id()] = idn
		wp.logger.Debug("WorkerPool: collected result", idn)
		return true
	} else {
		wp.logger.Debug("WorkerPool: result already collected, skip it", idn)
		return false
	}
}

// WaitFinalize waits until all tasks are processed and workers stopped
// and close the input/output channels
func (wp *workerPool) WaitFinalize() {
	wp.jobs.Wait()
	close(wp.tasksChan)
	wp.workers.Wait()
	wp.isFinalized = true
}

// Results gets final results or arises error if workers are still running
func (wp *workerPool) Results() ([]interface{}, error) {
	if !wp.isFinalized {
		err := fmt.Errorf("WorkerPool: workers are still running. Results not ready yet")
		wp.logger.Error(err.Error())
		return nil, err
	}

	results := make([]interface{}, 0)
	for _, v := range wp.results {
		results = append(results, v)
	}
	return results, nil
}
