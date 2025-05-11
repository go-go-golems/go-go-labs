package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Setup store and handlers
	store := NewStreamStore()
	h := NewStreamHandler(store)

	// Routes
	e.GET("/api/stream", h.GetStreamInfo)
	e.PUT("/api/stream", h.UpdateStreamInfo)
	e.GET("/api/stream/steps", h.GetSteps)
	e.PUT("/api/stream/steps/active", h.SetActiveStep)
	e.POST("/api/stream/steps/upcoming", h.AddUpcomingStep)
	e.POST("/api/stream/steps/complete", h.CompleteActiveStep)
	e.PUT("/api/stream/steps/reactivate", h.ReactivateStep)

	// Start server
	log.Println("Server started on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}