package helper

import (
	"fmt"
	"log/slog"
	"net/http"
)

type AppErrors struct {
	env    string
	logger *slog.Logger
}

func NewAppErrors(logger *slog.Logger, env string) *AppErrors {
	return &AppErrors{
		env:    env,
		logger: logger,
	}
}

// The logError() method is a generic helper for logging an error message along
// with the current request method and URL as attributes in the log entry.
func (ae *AppErrors) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	ae.logger.Error(err.Error(), "method", method, "uri", uri)
}

// The ErrorResponse() method is a generic helper for sending JSON-formatted error
// messages to the client with a given status code. Note that we're using the any
// type for the message parameter, rather than just a string type, as this gives us
// more flexibility over the values that we can include in the response.
func (ae *AppErrors) ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := Envelope{"error": message}

	// Write the response using the writeJSON() helper. If this happens to return an
	// error then log it, and fall back to sending the client an empty response with a
	// 500 Internal Server Error status code.
	err := WriteJSON(w, status, env, nil, ae.env)
	if err != nil {
		ae.logError(r, err)
		w.WriteHeader(500)
	}
}

// The ServerErrorResponse() method will be used when our application encounters an
// unexpected problem at runtime. It logs the detailed error message, then uses the
// errorResponse() helper to send a 500 Internal Server Error status code and JSON
// response (containing a generic error message) to the client.
func (ae *AppErrors) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	ae.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	ae.ErrorResponse(w, r, http.StatusInternalServerError, message)
}

// The NotFoundResponse() method will be used to send a 404 Not Found status code and
// JSON response to the client.
func (ae *AppErrors) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	ae.ErrorResponse(w, r, http.StatusNotFound, message)
}

// The MethodNotAllowedResponse() method will be used to send a 405 Method Not Allowed
// status code and JSON response to the client.
func (ae *AppErrors) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	ae.ErrorResponse(w, r, http.StatusMethodNotAllowed, message)
}
