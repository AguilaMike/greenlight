package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/handler"
)

type AppHandler struct {
	app        *config.Application
	apiVersion string
	areaName   string
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
	r.HandlerFunc(http.MethodGet, h.getURLPattern("healthcheck"), h.healthcheckHandler)
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (h *MainHandler) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", h.app.Config.Env)
	fmt.Fprintf(w, "version: %s\n", config.VERSION)
}
