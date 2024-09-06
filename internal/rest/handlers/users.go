package handlers

import (
	"errors"
	"net/http"

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

	// Call the Send() method on our Mailer, passing in the user's email address,
	// name of the template file, and the User struct containing the new user's data.
	err = uh.app.Mailer.Send(user.Email, "user_welcome.tmpl", user)
	if err != nil {
		uh.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// Write a JSON response containing the user data along with a 201 Created status
	// code.
	err = helper.WriteJSON(w, http.StatusCreated, helper.Envelope{"user": user}, nil, uh.app.Config.Env.String())
	if err != nil {
		uh.app.Errors.ServerErrorResponse(w, r, err)
	}
}
