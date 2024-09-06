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

type UserHandler struct {
	AppHandler
}

func NewUserHandler(app *config.Application) handler.AreaHandler {
	return &UserHandler{
		AppHandler: AppHandler{
			app:        app,
			apiVersion: config.API_VERSION,
			areaName:   "users",
		},
	}
}

func (u *UserHandler) SetRoutes(r *httprouter.Router) {
	r.HandlerFunc(http.MethodPost, u.getURLPattern(u.areaName), u.registerUserHandler)
	r.HandlerFunc(http.MethodPut, u.getURLPattern(u.areaName+"/activated"), u.activateUserHandler)
}

func (uh *UserHandler) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an anonymous struct to hold the expected data from the request body.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse the request body into the anonymous struct.
	err := helper.ReadJSON(w, r, &input)
	if err != nil {
		uh.app.Errors.BadRequestResponse(w, r, err)
		return
	}

	// Copy the data from the request body into a new User struct. Notice also that we
	// set the Activated field to false, which isn't strictly necessary because the
	// Activated field will have the zero-value of false by default. But setting this
	// explicitly helps to make our intentions clear to anyone reading the code.
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// Use the Password.Set() method to generate and store the hashed and plaintext
	// passwords.
	err = user.Password.Set(input.Password)
	if err != nil {
		uh.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	// Validate the user struct and return the error messages to the client if any of
	// the checks fail.
	if data.ValidateUser(v, user); !v.Valid() {
		uh.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user data into the database.
	err = uh.app.Models.Users.Insert(user)
	if err != nil {
		switch {
		// If we get a ErrDuplicateEmail error, use the v.AddError() method to manually
		// add a message to the validator instance, and then call our
		// failedValidationResponse() helper.
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			uh.app.Errors.FailedValidationResponse(w, r, v.Errors)
		default:
			uh.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Add the "movies:read" permission for the new user.
	err = uh.app.Models.Permissions.AddForUser(user.ID, permissionReadOnly)
	if err != nil {
		uh.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// After the user record has been created in the database, generate a new activation
	// token for the user.
	token, err := uh.app.Models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		uh.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// Launch a goroutine which runs an anonymous function that sends the welcome email.
	// Use the background helper to execute an anonymous function that sends the welcome email.
	uh.app.Worker.Background(func() {
		// As there are now multiple pieces of data that we want to pass to our email
		// templates, we create a map to act as a 'holding structure' for the data. This
		// contains the plaintext version of the activation token for the user, along
		// with their ID.
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		// Send the welcome email, passing in the map above as dynamic data.
		err = uh.app.Mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			// Importantly, if there is an error sending the email then we use the
			// app.logger.Error() helper to manage it, instead of the
			// app.serverErrorResponse() helper like before.
			uh.app.Logger.Error(err.Error())
		}
	})

	// Write a JSON response containing the user data along with a 201 Created status code.
	// Note that we also change this to send the client a 202 Accepted status code.
	// This status code indicates that the request has been accepted for processing, but
	// the processing has not been completed.
	err = helper.WriteJSON(w, http.StatusCreated, helper.Envelope{"user": user}, nil, uh.app.Config.Env.String())
	if err != nil {
		uh.app.Errors.ServerErrorResponse(w, r, err)
	}
}

func (uh *UserHandler) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the plaintext activation token from the request body.
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := helper.ReadJSON(w, r, &input)
	if err != nil {
		uh.app.Errors.BadRequestResponse(w, r, err)
		return
	}

	// Validate the plaintext token provided by the client.
	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		uh.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve the details of the user associated with the token using the
	// GetForToken() method (which we will create in a minute). If no matching record
	// is found, then we let the client know that the token they provided is not valid.
	user, err := uh.app.Models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			uh.app.Errors.FailedValidationResponse(w, r, v.Errors)
		default:
			uh.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Update the user's activation status.
	user.Activated = true

	// Save the updated user record in our database, checking for any edit conflicts in
	// the same way that we did for our movie records.
	err = uh.app.Models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			uh.app.Errors.EditConflictResponse(w, r)
		default:
			uh.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// If everything went successfully, then we delete all activation tokens for the
	// user.
	err = uh.app.Models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		uh.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// Send the updated user details to the client in a JSON response.
	err = helper.WriteJSON(w, http.StatusOK, helper.Envelope{"user": user}, nil, uh.app.Config.Env.String())
	if err != nil {
		uh.app.Errors.ServerErrorResponse(w, r, err)
	}
}
