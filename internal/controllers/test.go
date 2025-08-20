package controllers

import (
	"fmt"
	"log"
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
	ui, err := tc.test.Test()
	if err != nil {
		log.Printf("failed to load test data: %v", err)
		http.Error(w, "Test failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Test completed successfully, userID: %d", ui)
}
