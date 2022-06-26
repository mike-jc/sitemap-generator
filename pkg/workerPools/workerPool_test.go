package workerPools_test

import (
	"fmt"
	"os"
	"sitemap-generator/pkg/workerPools"
	"sitemap-generator/services"
	"sitemap-generator/utils"
	"strconv"
	"testing"
)

type idn int

func (i *idn) Id() string {
	return strconv.Itoa(int(*i))
}

func TestNewWorkerPool(t *testing.T) {
	logger, err := services.NewLogger(os.Stderr, "testing", "error")
	utils.AssertNoError(t, err)

	workersCount := 2
	tasks := []idn{3, 11, 33}
	expectedResults := []int{3, 11, 33, 4, 5, 34, 35}

	wp := workerPools.NewWorkerPool(logger, workersCount)

	var needsTask workerPools.NeedsTaskFunc = func(v interface{}) bool {
		if i, ok := v.(*idn); ok {
			ii := int(*i)
			// require task for number divisible by 3
			return ii%3 == 0
		}
		return false
	}
	var handler workerPools.WorkerHandler = func(v interface{}) error {
		if i, ok := v.(*idn); ok {
			ii := int(*i)
			// produce new input
			for j := 1; j <= 2; j++ {
				task := idn(ii + j)
				wp.Dispatch(&task, needsTask)
			}
			return nil
		}
		return fmt.Errorf("wrong type of handler's input: %T", v)
	}

	startedWorkers, wpErr := wp.Init(handler)
	utils.AssertNoError(t, wpErr)
	utils.AssertEqual(t, startedWorkers, workersCount)

	for _, task := range tasks {
		func(task idn) {
			wp.Dispatch(&task, needsTask)
		}(task)
	}

	wp.WaitFinalize()

	results := make([]int, 0)
	wpResults, _ := wp.Results()

	for _, r := range wpResults {
		if i, ok := r.(*idn); ok {
			ii := int(*i)
			results = append(results, ii)
		} else {
			t.Errorf("wrong type of worker pool's result: %T", r)
		}
	}
	utils.AssertEqualSlices(t, results, expectedResults)
}
