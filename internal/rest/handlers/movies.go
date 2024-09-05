package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/internal/data"
	"github.com/AguilaMike/greenlight/internal/validator"
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
	r.HandlerFunc(http.MethodGet, m.getURLPattern(m.areaName), m.listMoviesHandler)
	r.HandlerFunc(http.MethodPost, m.getURLPattern(m.areaName), m.createMovieHandler)
	r.HandlerFunc(http.MethodGet, m.getURLPattern(m.areaName+"/:id"), m.showMovieHandler)
	r.HandlerFunc(http.MethodPatch, m.getURLPattern(m.areaName+"/:id"), m.updateMovieHandler)
	r.HandlerFunc(http.MethodDelete, m.getURLPattern(m.areaName+"/:id"), m.deleteMovieHandler)
}

func (m *MovieHandler) getPayloadFromRequest(w http.ResponseWriter, r *http.Request, movie *data.Movie, requiredAll bool) (succes, hasChanged bool) {
	hasChanged = requiredAll
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body (note that the field names and types in the struct are a subset
	// of the Movie struct that we created earlier). This struct will be our *target
	// decode destination*.
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := helper.ReadJSON(w, r, &input)
	if err != nil {
		m.app.Errors.BadRequestResponse(w, r, err)
		return false, false
	}

	// Copy the values from the request body to the appropriate fields of the movie
	// record. We only want to update the fields in the movie record if they have been
	// provided in the request body, so we use the nil coalescing operator to check if
	// the fields in the input struct are nil. If they're not, we update the corresponding
	// field in the movie record.
	if input.Title != nil || requiredAll {
		movie.Title = *input.Title
		hasChanged = true
	}
	if input.Year != nil || requiredAll {
		movie.Year = *input.Year
		hasChanged = true
	}
	if input.Runtime != nil || requiredAll {
		movie.Runtime = *input.Runtime
		hasChanged = true
	}
	if input.Genres != nil || requiredAll {
		movie.Genres = input.Genres
		hasChanged = true
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the ValidateMovie() function and return a response containing the errors if
	// any of the checks fail.
	// Use the Valid() method to see if any of the checks failed. If they did, then use
	// the failedValidationResponse() helper to send a response to the client, passing
	// in the v.Errors map.
	if data.ValidateMovie(v, movie); !v.Valid() {
		m.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return false, false
	}

	return true, hasChanged
}

func (m *MovieHandler) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string.
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	// Use our helpers to extract the title and genres query string values, falling back
	// to defaults of an empty string and an empty slice respectively if they are not
	// provided by the client.
	input.Title = helper.QpReadString(qs, "title", "")
	input.Genres = helper.QpReadCSV(qs, "genres", []string{})

	// Get the page and page_size query string values as integers. Notice that we set
	// the default page value to 1 and default page_size to 20, and that we pass the
	// validator instance as the final argument here.
	input.Filters.Page = helper.QpReadInt(qs, "page", 1, v)
	input.Filters.PageSize = helper.QpReadInt(qs, "page_size", 20, v)

	// Extract the sort query string value, falling back to "id" if it is not provided
	// by the client (which will imply a ascending sort on movie ID).
	input.Filters.Sort = helper.QpReadString(qs, "sort", "id")

	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// Execute the validation checks on the Filters struct and send a response
	// containing the errors if necessary.
	// Check the Validator instance for any errors and use the failedValidationResponse()
	// helper to send the client a response if necessary.
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		m.app.Errors.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the GetAll() method to retrieve the movies, passing in the various filter
	// parameters.
	movies, err := m.app.Models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		m.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the movie data.
	err = helper.WriteJSON(w, http.StatusOK, helper.Envelope{"movies": movies}, nil, m.app.Config.Env.String())
	if err != nil {
		m.app.Errors.ServerErrorResponse(w, r, err)
	}
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

	// Call the Get() method to fetch the data for a specific movie. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	movie, err := m.app.Models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			m.app.Errors.NotFoundResponse(w, r)
		default:
			m.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = helper.WriteJSON(w, http.StatusOK, helper.Envelope{"movie": movie}, nil, m.app.Config.Env.String())
	if err != nil {
		m.app.Errors.ServerErrorResponse(w, r, err)
	}
}

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.
func (m *MovieHandler) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	movie := &data.Movie{}
	if success, _ := m.getPayloadFromRequest(w, r, movie, true); !success {
		return
	}

	// Call the Insert() method on our movies model, passing in a pointer to the
	// validated movie struct. This will create a record in the database and update the
	// movie struct with the system-generated information.
	err := m.app.Models.Movies.Insert(movie)
	if err != nil {
		m.app.Errors.ServerErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = helper.WriteJSON(w, http.StatusCreated, helper.Envelope{"movie": movie}, headers, m.app.Config.Env.String())
	if err != nil {
		m.app.Errors.ServerErrorResponse(w, r, err)
	}
}

func (m *MovieHandler) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := helper.ReadParamFromRequest[int64](r, "id")
	if err != nil || id < 1 {
		m.app.Errors.NotFoundResponse(w, r)
		return
	}

	// Fetch the existing movie record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	movie, err := m.app.Models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			m.app.Errors.NotFoundResponse(w, r)
		default:
			m.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	sucess, hasChanged := m.getPayloadFromRequest(w, r, movie, false)
	if !sucess {
		return
	}

	if !hasChanged {
		err = helper.WriteJSON(w, http.StatusOK, helper.Envelope{"movie": movie}, nil, m.app.Config.Env.String())
		if err != nil {
			m.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Pass the updated movie record to our new Update() method.
	// Intercept any ErrEditConflict error and call the new editConflictResponse()
	// helper.
	err = m.app.Models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			m.app.Errors.EditConflictResponse(w, r)
		default:
			m.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Write the updated movie record in a JSON response.
	err = helper.WriteJSON(w, http.StatusOK, helper.Envelope{"movie": movie}, nil, m.app.Config.Env.String())
	if err != nil {
		m.app.Errors.ServerErrorResponse(w, r, err)
	}
}

func (m *MovieHandler) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := helper.ReadParamFromRequest[int64](r, "id")
	if err != nil || id < 1 {
		m.app.Errors.NotFoundResponse(w, r)
		return
	}

	// Delete the movie from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = m.app.Models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			m.app.Errors.NotFoundResponse(w, r)
		default:
			m.app.Errors.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status code along with a success message.
	err = helper.WriteJSON(w, http.StatusOK, helper.Envelope{"message": "movie successfully deleted"}, nil, m.app.Config.Env.String())
	if err != nil {
		m.app.Errors.ServerErrorResponse(w, r, err)
	}
}
