package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/internal/data"
	"github.com/AguilaMike/greenlight/internal/validator"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/handler"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/helper"
)

type TokenHandler struct {
	AppHandler
}

func NewTokenHandler(app *config.Application) handler.AreaHandler {
	return &TokenHandler{
		AppHandler: AppHandler{
			app:        app,
			apiVersion: config.API_VERSION,
			areaName:   "tokens",
		},
	}
}

func (u *TokenHandler) SetRoutes(r *httprouter.Router) {
	r.HandlerFunc(http.MethodPost, u.getURLPattern(u.areaName)+"/authentication", u.createAuthenticationTokenHandler)
}

func (th *TokenHandler) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the email and password from the request body.
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := helper.ReadJSON(w, r, &input)
	if err != nil {
		th.app.Errors.BadRequestResponse(w, r, err)
		return
	}

	// Validate the email and password provided by the client.
	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		th.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Lookup the user record based on the email address. If no matching user was
	// found, then we call the app.invalidCredentialsResponse() helper to send a 401
	// Unauthorized response to the client (we will create this helper in a moment).
	user, err := th.app.Models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			th.app.Errors.InvalidCredentialsResponse(w, r)
		default:
			th.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Check if the provided password matches the actual password for the user.
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		th.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// If the passwords don't match, then we call the app.invalidCredentialsResponse()
	// helper again and return.
	if !match {
		th.app.Errors.InvalidCredentialsResponse(w, r)
		return
	}

	// Otherwise, if the password is correct, we generate a new token with a 24-hour
	// expiry time and the scope 'authentication'.
	token, err := th.app.Models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		th.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// Encode the token to JSON and send it in the response along with a 201 Created
	// status code.
	err = helper.WriteJSON(w, http.StatusCreated, helper.Envelope{"authentication_token": token}, nil, th.app.Config.Env.String())
	if err != nil {
		th.app.Errors.ServerErrorResponse(w, r, err)
	}
}
