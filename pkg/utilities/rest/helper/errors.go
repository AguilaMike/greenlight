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

/******************************************
* 400 Bad Request Response Helper Methods *
******************************************/
// The BadRequestResponse() method will be used to send a 400 Bad Request status code
// and JSON response to the client.
// 400 Bad Request Response Helper Method
func (ae *AppErrors) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	ae.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}

// The ValidationResponse() method will be used to send a 400 Bad Request status code
// and JSON response to the client, when the request body fails validation.
// 401 Unauthorized Response Helper Method
func (app *AppErrors) InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.ErrorResponse(w, r, http.StatusUnauthorized, message)
}

// The InvalidAuthenticationTokenResponse() method will be used to send a 401 Unauthorized
// status code and JSON response to the client.
// 401 Unauthorized Response Helper Method
func (app *AppErrors) InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	app.ErrorResponse(w, r, http.StatusUnauthorized, message)
}

// The authenticationRequiredResponse() method will be used to send a 401 Unauthorized
// status code and JSON response to the client.
// 401 Unauthorized Response Helper Method
func (app *AppErrors) AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.ErrorResponse(w, r, http.StatusUnauthorized, message)
}

// The forbiddenResponse() method will be used to send a 403 Forbidden status code and
// JSON response to the client.
// 403 Forbidden Response Helper Method
func (app *AppErrors) InactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	app.ErrorResponse(w, r, http.StatusForbidden, message)
}

// The NotFoundResponse() method will be used to send a 404 Not Found status code and
// JSON response to the client.
// 404 Not Found Response Helper Method
func (ae *AppErrors) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	ae.ErrorResponse(w, r, http.StatusNotFound, message)
}

// The MethodNotAllowedResponse() method will be used to send a 405 Method Not Allowed
// status code and JSON response to the client.
// 405 Method Not Allowed Response Helper Method
func (ae *AppErrors) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	ae.ErrorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// The EditConflictResponse() method will be used to send a 409 Conflict status code and
// JSON response to the client.
// 409 Conflict Response Helper Method
func (ae *AppErrors) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	ae.ErrorResponse(w, r, http.StatusConflict, message)
}

// Note that the errors parameter here has the type map[string]string, which is exactly
// the same as the errors map contained in our Validator type.
// 422 Unprocessable Entity Response Helper Method
func (ae *AppErrors) FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	ae.ErrorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// The RateLimitExceededResponse() method will be used to send a 429 Too Many Requests
// status code and JSON response to the client.
// 429 Too Many Requests Response Helper Method
func (ae *AppErrors) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	ae.ErrorResponse(w, r, http.StatusTooManyRequests, message)
}

/****************************************************
* 500 Internal Server Error Response Helper Methods *
****************************************************/
// The ServerErrorResponse() method will be used when our application encounters an
// unexpected problem at runtime. It logs the detailed error message, then uses the
// errorResponse() helper to send a 500 Internal Server Error status code and JSON
// response (containing a generic error message) to the client.
// 500 Internal Server Error Response Helper Method
func (ae *AppErrors) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	ae.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	ae.ErrorResponse(w, r, http.StatusInternalServerError, message)
}
