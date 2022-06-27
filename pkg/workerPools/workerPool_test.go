package workerPools_test

import (
	"fmt"
	"os"
	"sitemap-generator/pkg/workerPools"
	"sitemap-generator/services"
	"sitemap-generator/utils"
	"testing"
)

func TestNewWorkerPool(t *testing.T) {
	logger, err := services.NewLogger(os.Stderr, "testing", "error")
	utils.AssertNoError(t, err)

	workersCount := 2
	tasks := []int{3, 10, 32}
	expectedResults := []int{3, 10, 32, 12, 33}

	results := make([]int, 0)
	wp := workerPools.NewWorkerPool(logger, workersCount)

	var handler workerPools.WorkerHandler = func(v interface{}) error {
		if i, ok := v.(int); ok {
			// collect result
			results = append(results, i)

			// produce new input and decide if new task is needed
			for j := 1; j <= 2; j++ {
				task := i + j
				if task%3 == 0 {
					wp.AddTask(task)
				}
			}
			return nil
		}
		return fmt.Errorf("wrong type of handler's input: %T", v)
	}

	startedWorkers, wpErr := wp.Init(handler)
	utils.AssertNoError(t, wpErr)
	utils.AssertEqual(t, startedWorkers, workersCount)

	// initial iteration
	for _, t := range tasks {
		_ = handler(t)
	}
	wp.WaitFinalize()
	utils.AssertEqualSlices(t, results, expectedResults)
}
