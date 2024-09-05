package routes

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/internal/rest/handlers"
)

func GenerateRoutes(cfg *config.Application) http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	// Create routes for the main handler.
	handlers.NewMainHandler(cfg).SetRoutes(router)

	// Create routes for the movie handler.
	handlers.NewMovieHandler(cfg).SetRoutes(router)

	// Return the httprouter instance.
	return router
}
