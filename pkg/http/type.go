package http

import (
	"net/http"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/hashicorp/go-hclog"
)

// Server wraps up all the request routers and associated components
// that serve various parts of the nbuild stack.
type Server struct {
	l hclog.Logger
	r chi.Router

	n     *http.Server
	f     *form.Decoder
	tmpls *pongo2.TemplateSet
}

// Team is a team object that may be checked in or have other distinct
// status.  Once loaded the number cannot be changed.
type Team struct {
	Number int
	Name   string
}
