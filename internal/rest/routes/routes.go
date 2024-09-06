package routes

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/internal/rest/handlers"
	"github.com/AguilaMike/greenlight/internal/rest/middlewares"
)

func GenerateRoutes(cfg *config.Application) http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	// Initialize a new Middleware instance.
	middleware := middlewares.NewAppMiddleware(cfg)

	// Create routes for the main handler.
	handlers.NewMainHandler(cfg).SetRoutes(router)

	// Create routes for the movie handler.
	handlers.NewMovieHandler(cfg).SetRoutes(router)

	// Create routes for the user handler.
	handlers.NewUserHandler(cfg).SetRoutes(router)

	// Create routes for the token handler.
	handlers.NewTokenHandler(cfg).SetRoutes(router)

	// Return the httprouter instance.
	return middleware.RateLimit(middleware.RecoverPanic(router))
}
