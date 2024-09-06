package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/internal/rest/middlewares"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/handler"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/helper"
)

type AppHandler struct {
	app        *config.Application
	apiVersion string
	areaName   string
	mid        *middlewares.AppMiddleware
}

func (ah *AppHandler) GetAreaName() string {
	return ah.areaName
}

func (ah *AppHandler) SetRoutes(r *httprouter.Router) {
	panic("not implemented")
}

func (ah *AppHandler) getURLPattern(url string) string {
	return fmt.Sprintf("/%s/%s", ah.apiVersion, url)
}

type MainHandler struct {
	AppHandler
}

func NewMainHandler(app *config.Application) handler.AreaHandler {
	return &MainHandler{
		AppHandler: AppHandler{
			app:        app,
			apiVersion: config.API_VERSION,
			areaName:   "api",
		},
	}
}

func (h *MainHandler) SetRoutes(r *httprouter.Router) {
	// Paths for general API application
	r.HandlerFunc(http.MethodGet, h.getURLPattern("healthcheck"), h.healthcheckHandler)

	// Paths for manage errors

	// Convert the notFoundResponse() helper to a http.Handler using the
	// http.HandlerFunc() adapter, and then set it as the custom error handler for 404
	// Not Found responses.
	r.NotFound = http.HandlerFunc(h.app.Errors.NotFoundResponse)

	// Likewise, convert the methodNotAllowedResponse() helper to a http.Handler and set
	// it as the custom error handler for 405 Method Not Allowed responses.
	r.MethodNotAllowed = http.HandlerFunc(h.app.Errors.MethodNotAllowedResponse)
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (h *MainHandler) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map which holds the information that we want to send in the response.
	data := helper.Envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": h.app.Config.Env.String(),
			"version":     config.VERSION,
		},
	}

	// Call the WriteJSON() helper, passing in the http.ResponseWriter, the map, and nil for the
	// header map. If the helper returns an error, log the detailed error message and send a
	// generic error response to the client.
	err := helper.WriteJSON(w, http.StatusOK, data, nil, h.app.Config.Env.String())
	if err != nil {
		h.app.Errors.ServerErrorResponse(w, r, err)
	}
}
