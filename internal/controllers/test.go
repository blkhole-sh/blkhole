package controllers

import (
	"fmt"
	"net/http"
	"github.com/lemon3studio/leo/internal/test"
)

// TestController defines the interface for test operations
type TestController interface {
	RunTest(w http.ResponseWriter, r *http.Request)
}

// testController implements the TestController interface
type testController struct {
	test test.Test
}

// NewTestController creates a new TestController instance
func NewTestController(test test.Test) TestController {
	return &testController{
		test: test,
	}
}

// RunTest executes the test suite
func (tc *testController) RunTest(w http.ResponseWriter, r *http.Request) {
	uh, err := tc.test.Test()
	if err != nil {
		http.Error(w, "Test failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Test completed successfully, userHash: %s", uh)
}
