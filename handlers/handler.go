package handlers

import "arturgudiev/memoryguard/app"

// Handler holds application dependencies.
type Handler struct {
	App *app.App
}

// NewHandler creates a new handler instance.
func NewHandler(application *app.App) *Handler {
	return &Handler{App: application}
}
