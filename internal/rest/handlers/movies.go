package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/internal/data"
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

// Add a showMovieHandler for the "GET /v1/movies/:id" endpoint. For now, we retrieve
// the interpolated "id" parameter from the current URL and include it in a placeholder
// response.
func (m *MovieHandler) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ReadParamFromRequest[int64](r, "id")
	if err != nil || id < 1 {
		m.app.Errors.NotFoundResponse(w, r)
		return
	}

	// Create a new instance of the Movie struct, containing the ID we extracted from
	// the URL and some dummy data. Also notice that we deliberately haven't set a
	// value for the Year field.
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = helper.WriteJSON(w, http.StatusOK, helper.Envelope{"movie": movie}, nil, m.app.Config.Env.String())
	if err != nil {
		m.app.Errors.ServerErrorResponse(w, r, err)
	}
}

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.
func (m *MovieHandler) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}
