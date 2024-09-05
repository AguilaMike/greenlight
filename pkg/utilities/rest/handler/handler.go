package handler

import "github.com/julienschmidt/httprouter"

type AreaHandler interface {
	GetAreaName() string
	SetRoutes(r *httprouter.Router)
}
