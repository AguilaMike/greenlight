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
	r.HandlerFunc(http.MethodPost, u.getURLPattern(u.areaName)+"/activation", u.createActivationTokenHandler)
	r.HandlerFunc(http.MethodPost, u.getURLPattern(u.areaName)+"/password-reset", u.createPasswordResetTokenHandler)
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

// Generate a password reset token and send it to the user's email address.
func (th *TokenHandler) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate the user's email address.
	var input struct {
		Email string `json:"email"`
	}

	err := helper.ReadJSON(w, r, &input)
	if err != nil {
		th.app.Errors.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		th.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Try to retrieve the corresponding user record for the email address. If it can't
	// be found, return an error message to the client.
	user, err := th.app.Models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			th.app.Errors.FailedValidationResponse(w, r, v.Errors)
		default:
			th.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Return an error message if the user is not activated.
	if !user.Activated {
		v.AddError("email", "user account must be activated")
		th.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Otherwise, create a new password reset token with a 45-minute expiry time.
	token, err := th.app.Models.Tokens.New(user.ID, 45*time.Minute, data.ScopePasswordReset)
	if err != nil {
		th.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// Email the user with their password reset token.
	th.app.Worker.Background(func() {
		data := map[string]any{
			"passwordResetToken": token.Plaintext,
		}

		// Since email addresses MAY be case sensitive, notice that we are sending this
		// email using the address stored in our database for the user --- not to the
		// input.Email address provided by the client in this request.
		err = th.app.Mailer.Send(user.Email, "token_password_reset.tmpl", data)
		if err != nil {
			th.app.Logger.Error(err.Error())
		}
	})

	// Send a 202 Accepted response and confirmation message to the client.
	env := helper.Envelope{"message": "an email will be sent to you containing password reset instructions"}

	err = helper.WriteJSON(w, http.StatusAccepted, env, nil, th.app.Config.Env.String())
	if err != nil {
		th.app.Errors.ServerErrorResponse(w, r, err)
	}
}

func (th *TokenHandler) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate the user's email address.
	var input struct {
		Email string `json:"email"`
	}

	err := helper.ReadJSON(w, r, &input)
	if err != nil {
		th.app.Errors.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		th.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Try to retrieve the corresponding user record for the email address. If it can't
	// be found, return an error message to the client.
	user, err := th.app.Models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			th.app.Errors.FailedValidationResponse(w, r, v.Errors)
		default:
			th.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Return an error if the user has already been activated.
	if user.Activated {
		v.AddError("email", "user has already been activated")
		th.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Otherwise, create a new activation token.
	token, err := th.app.Models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		th.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// Email the user with their additional activation token.
	th.app.Worker.Background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
		}

		// Since email addresses MAY be case sensitive, notice that we are sending this
		// email using the address stored in our database for the user --- not to the
		// input.Email address provided by the client in this request.
		err = th.app.Mailer.Send(user.Email, "token_activation.tmpl", data)
		if err != nil {
			th.app.Logger.Error(err.Error())
		}
	})

	// Send a 202 Accepted response and confirmation message to the client.
	env := helper.Envelope{"message": "an email will be sent to you containing activation instructions"}

	err = helper.WriteJSON(w, http.StatusAccepted, env, nil, th.app.Config.Env.String())
	if err != nil {
		th.app.Errors.ServerErrorResponse(w, r, err)
	}
}
