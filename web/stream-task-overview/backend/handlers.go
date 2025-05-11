package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// StreamHandler handles HTTP requests
type StreamHandler struct {
	store *StreamStore
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(store *StreamStore) *StreamHandler {
	log.Debug().Msg("Creating new StreamHandler")
	return &StreamHandler{store: store}
}

// GetStreamInfo returns stream information
func (h *StreamHandler) GetStreamInfo(c echo.Context) error {
	log.Debug().Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).Msg("Handling GetStreamInfo request")
	
	info := h.store.GetStreamInfo()
	log.Debug().Interface("info", info).Msg("Returning stream info")
	
	return c.JSON(http.StatusOK, info)
}

// UpdateStreamInfo updates stream information
func (h *StreamHandler) UpdateStreamInfo(c echo.Context) error {
	log.Debug().Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).Msg("Handling UpdateStreamInfo request")
	
	var info StreamInfo
	if err := c.Bind(&info); err != nil {
		log.Error().Err(err).Msg("Failed to parse UpdateStreamInfo request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}
	
	log.Debug().Interface("info", info).Msg("Updating stream info")
	h.store.UpdateStreamInfo(info)
	
	updatedInfo := h.store.GetStreamInfo()
	log.Debug().Interface("info", updatedInfo).Msg("Returning updated stream info")
	
	return c.JSON(http.StatusOK, updatedInfo)
}

// GetSteps returns all steps
func (h *StreamHandler) GetSteps(c echo.Context) error {
	log.Debug().Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).Msg("Handling GetSteps request")
	
	steps := h.store.GetSteps()
	log.Debug().Interface("steps", steps).Msg("Returning steps")
	
	return c.JSON(http.StatusOK, steps)
}

// SetActiveStep sets a new active step
func (h *StreamHandler) SetActiveStep(c echo.Context) error {
	log.Debug().Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).Msg("Handling SetActiveStep request")
	
	var data struct {
		Step string `json:"step"`
	}
	
	if err := c.Bind(&data); err != nil {
		log.Error().Err(err).Msg("Failed to parse SetActiveStep request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}
	
	log.Debug().Str("step", data.Step).Msg("Setting active step")
	h.store.SetActiveStep(data.Step)
	
	updatedSteps := h.store.GetSteps()
	log.Debug().Interface("steps", updatedSteps).Msg("Returning updated steps")
	
	return c.JSON(http.StatusOK, updatedSteps)
}

// AddUpcomingStep adds a new upcoming step
func (h *StreamHandler) AddUpcomingStep(c echo.Context) error {
	log.Debug().Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).Msg("Handling AddUpcomingStep request")
	
	var data struct {
		Step string `json:"step"`
	}
	
	if err := c.Bind(&data); err != nil {
		log.Error().Err(err).Msg("Failed to parse AddUpcomingStep request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}
	
	log.Debug().Str("step", data.Step).Msg("Adding upcoming step")
	h.store.AddUpcomingStep(data.Step)
	
	updatedSteps := h.store.GetSteps()
	log.Debug().Interface("steps", updatedSteps).Msg("Returning updated steps")
	
	return c.JSON(http.StatusOK, updatedSteps)
}

// CompleteActiveStep completes the current active step
func (h *StreamHandler) CompleteActiveStep(c echo.Context) error {
	log.Debug().Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).Msg("Handling CompleteActiveStep request")
	
	log.Debug().Msg("Completing active step")
	h.store.CompleteActiveStep()
	
	updatedSteps := h.store.GetSteps()
	log.Debug().Interface("steps", updatedSteps).Msg("Returning updated steps")
	
	return c.JSON(http.StatusOK, updatedSteps)
}

// ReactivateStep moves a step from completed/upcoming to active
func (h *StreamHandler) ReactivateStep(c echo.Context) error {
	log.Debug().Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).Msg("Handling ReactivateStep request")
	
	var data struct {
		Step   string `json:"step"`
		Source string `json:"source"` // "completed" or "upcoming"
	}
	
	if err := c.Bind(&data); err != nil {
		log.Error().Err(err).Msg("Failed to parse ReactivateStep request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}
	
	log.Debug().Str("step", data.Step).Str("source", data.Source).Msg("Reactivating step")
	h.store.ReactivateStep(data.Step, data.Source)
	
	updatedSteps := h.store.GetSteps()
	log.Debug().Interface("steps", updatedSteps).Msg("Returning updated steps")
	
	return c.JSON(http.StatusOK, updatedSteps)
}