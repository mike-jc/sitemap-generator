package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("%s\n%s\nError: %s", callerDetails(), t.Name(), err.Error())
	}
}

func AssertHasError(t *testing.T, err error, expectedText string) {
	if err == nil {
		t.Errorf("%s\n%s\nExpected error but got no error", callerDetails(), t.Name())
		return
	}
	if !strings.Contains(err.Error(), expectedText) {
		t.Errorf("%s\n%s\nError does not contain expected text.\nExpected text: %s\nActual error: %s", callerDetails(), t.Name(), expectedText, err.Error())
	}
}

func AssertEqual(t *testing.T, actual interface{}, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%s\n%s\nActual result is not as expected.\nExpected: %s\nActual: %s", callerDetails(), t.Name(), InJSON(expected), InJSON(actual))
	}
}

// AssertEqualSlices compares slices without checking order of elements
func AssertEqualSlices(t *testing.T, actual interface{}, expected interface{}) {
	if reflect.TypeOf(actual).Kind() != reflect.Slice {
		t.Errorf("%s\n%s\nWrang type of actual result: %T. Must be slice", callerDetails(), t.Name(), actual)
		return
	}
	if reflect.TypeOf(expected).Kind() != reflect.Slice {
		t.Errorf("%s\n%s\nWrang type of expected result %T. Must be slice", callerDetails(), t.Name(), expected)
		return
	}

	actualSlice := reflect.ValueOf(actual)
	expectedSlice := reflect.ValueOf(expected)

	equal := false
	if actualSlice.Len() == expectedSlice.Len() {
		equal = true
		for i := 0; i < expectedSlice.Len(); i++ {
			found := false
			elem := expectedSlice.Index(i).Interface()

			for j := 0; j < actualSlice.Len(); j++ {
				if reflect.DeepEqual(elem, actualSlice.Index(j).Interface()) {
					found = true
					break
				}
			}
			if !found {
				equal = false
				break
			}
		}
	}

	if !equal {
		t.Errorf("%s\n%s\nActual result is not as expected.\nExpected: %s\nActual: %s", callerDetails(), t.Name(), InJSON(expected), InJSON(actual))
	}
}

func AssertEmpty(t *testing.T, v interface{}) {
	if !reflect.ValueOf(v).IsZero() {
		t.Errorf("%s\n%s\nNot empty as expected: %s", callerDetails(), t.Name(), InJSON(v))
	}
}

func AssertTrue(t *testing.T, v bool) {
	if !v {
		t.Errorf("%s\n%s\nNot true as expected", callerDetails(), t.Name())
	}
}

func AssertFalse(t *testing.T, v bool) {
	if v {
		t.Errorf("%s\n%s\nNot false as expected", callerDetails(), t.Name())
	}
}

func callerDetails() string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}
