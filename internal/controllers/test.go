package controllers

import (
	"fmt"
	"net/http"
	"server/internal/test"
)

// TestController defines the interface for test operations
type TestController interface {
	RunTest(w http.ResponseWriter, r *http.Request)
}

// TestControllerImpl implements the TestController interface
type TestControllerImpl struct {
	test test.Test
}

// NewTestController creates a new TestController instance
func NewTestController(test test.Test) TestController {
	return &TestControllerImpl{
		test: test,
	}
}

// RunTest executes the test suite
func (tc *TestControllerImpl) RunTest(w http.ResponseWriter, r *http.Request) {
	uh, err := tc.test.Test()
	if err != nil {
		http.Error(w, "Test failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Test completed successfully, userHash: %s", uh)
}
