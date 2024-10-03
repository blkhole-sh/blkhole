package controllers

import (
	"net/http"
	"server/services"
	"strconv"
)

// Define ScheduleController struct
type ScheduleController struct {
	ContentBlocker *services.ContentBlocker
}

// Create new ScheduleController
func NewScheduleController(contentBlocker *services.ContentBlocker) *ScheduleController {
	return &ScheduleController{ContentBlocker: contentBlocker}
}

// GET request to check if a domain given as request param is blocked
func (sc *ScheduleController) IsBlocked(w http.ResponseWriter, r *http.Request) {
	// Parse domain from request param
	domain := r.URL.Query().Get("domain")

	// Check if domain is blocked, throw error if domain is invalid
	blocked, err := sc.ContentBlocker.IsBlocked(domain)
	if err != nil {

		// Return HTTP Bad Gateway if domain is invalid
		http.Error(w, "Invalid domain", http.StatusBadGateway)
		return
	}

	// Return HTTP 200 and true if domain is blocked or false if else
	w.Write([]byte(strconv.FormatBool(blocked)))
}
