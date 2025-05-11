package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// StreamHandler handles HTTP requests
type StreamHandler struct {
	store *StreamStore
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(store *StreamStore) *StreamHandler {
	return &StreamHandler{store: store}
}

// GetStreamInfo returns stream information
func (h *StreamHandler) GetStreamInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, h.store.GetStreamInfo())
}

// UpdateStreamInfo updates stream information
func (h *StreamHandler) UpdateStreamInfo(c echo.Context) error {
	var info StreamInfo
	if err := c.Bind(&info); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	h.store.UpdateStreamInfo(info)
	return c.JSON(http.StatusOK, h.store.GetStreamInfo())
}

// GetSteps returns all steps
func (h *StreamHandler) GetSteps(c echo.Context) error {
	return c.JSON(http.StatusOK, h.store.GetSteps())
}

// SetActiveStep sets a new active step
func (h *StreamHandler) SetActiveStep(c echo.Context) error {
	var data struct {
		Step string `json:"step"`
	}
	
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	h.store.SetActiveStep(data.Step)
	return c.JSON(http.StatusOK, h.store.GetSteps())
}

// AddUpcomingStep adds a new upcoming step
func (h *StreamHandler) AddUpcomingStep(c echo.Context) error {
	var data struct {
		Step string `json:"step"`
	}
	
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	h.store.AddUpcomingStep(data.Step)
	return c.JSON(http.StatusOK, h.store.GetSteps())
}

// CompleteActiveStep completes the current active step
func (h *StreamHandler) CompleteActiveStep(c echo.Context) error {
	h.store.CompleteActiveStep()
	return c.JSON(http.StatusOK, h.store.GetSteps())
}

// ReactivateStep moves a step from completed/upcoming to active
func (h *StreamHandler) ReactivateStep(c echo.Context) error {
	var data struct {
		Step   string `json:"step"`
		Source string `json:"source"` // "completed" or "upcoming"
	}
	
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	h.store.ReactivateStep(data.Step, data.Source)
	return c.JSON(http.StatusOK, h.store.GetSteps())
}