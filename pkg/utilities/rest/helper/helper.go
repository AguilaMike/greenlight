package helper

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Kind is an enum type for kind
type Kind int

const (
	KindString Kind = iota
	KindInt
)

func ReadParamFromRequest[T any](r *http.Request, key string, kind Kind) (T, error) {
	// When httprouter is parsing a request, any interpolated URL parameters will be
	// stored in the request context. We can use the ParamsFromContext() function to
	// retrieve a slice containing these parameter names and values.
	params := httprouter.ParamsFromContext(r.Context())
	// We can then use the ByName() method to get the value of the "id" parameter from
	// the slice. In our project all movies will have a unique positive integer ID, but
	// the value returned by ByName() is always a string. So we try to convert it to a
	// base 10 integer (with a bit size of 64). If the parameter couldn't be converted,
	// or is less than 1, we know the ID is invalid so we use the http.NotFound()
	// function to return a 404 Not Found response.
	param := params.ByName(key)

	var result T

	switch kind {
	case KindString:
		if v, ok := any(param).(T); ok {
			return v, nil
		}
		return result, fmt.Errorf("type assertion to string failed")
	case KindInt:
		i, err := strconv.Atoi(param)
		if err != nil {
			return result, err
		}
		if v, ok := any(i).(T); ok {
			return v, nil
		}
		return result, fmt.Errorf("type assertion to int failed")
	default:
		return result, fmt.Errorf("unsupported kind %v", kind)
	}
}
