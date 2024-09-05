package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/handler"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/helper"
)

type MovieHandler struct {
	AppHandler
}

func NewMovieHandler(app *config.Application) handler.AreaHandler {
	return &MovieHandler{
		AppHandler: AppHandler{
			app:        app,
			apiVersion: config.API_VERSION,
			areaName:   "movies",
		},
	}
}

func (m *MovieHandler) SetRoutes(r *httprouter.Router) {
	r.HandlerFunc(http.MethodGet, m.getURLPattern(m.areaName+"/:id"), m.showMovieHandler)
	r.HandlerFunc(http.MethodPost, m.getURLPattern(m.areaName), m.createMovieHandler)
}

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.
func (app *MovieHandler) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// Add a showMovieHandler for the "GET /v1/movies/:id" endpoint. For now, we retrieve
// the interpolated "id" parameter from the current URL and include it in a placeholder
// response.
func (app *MovieHandler) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParamFromRequest[int](r, "id", helper.KindInt)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	// Otherwise, interpolate the movie ID in a placeholder response.
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}
