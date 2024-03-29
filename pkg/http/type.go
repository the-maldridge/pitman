package http

import (
	"context"
	"net/http"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/hashicorp/go-hclog"
)

// KV implementations provide a storage interface to a persistent
// location either on disk or on a remote location.
type KV interface {
	Keys(context.Context, string) ([]string, error)
	Get(context.Context, string) ([]byte, error)
	Put(context.Context, string, []byte) error
	Ping(context.Context) error
	Close() error
}

// Server wraps up all the request routers and associated components
// that serve various parts of the nbuild stack.
type Server struct {
	l hclog.Logger
	r chi.Router

	n  *http.Server
	f  *form.Decoder
	kv KV

	forms map[string]Form
	tmpls *pongo2.TemplateSet
}

// Team is a team object that may be checked in or have other distinct
// status.  Everything is a string because internally nothing uses the
// number as a number.
type Team struct {
	Hub    string
	Name   string
	Number string
	Table  string
}

// Form is the dynamically loaded compliance form, but is generic
// enough it could be extended for future forms if desired.
type Form struct {
	Title    string
	Sections []FormSection
}

// FormSection is a container for one or more groups of fields.
type FormSection struct {
	Label  string
	Groups []FormGroup
}

// FormGroup is a container for one or more fields with an associated
// label.
type FormGroup struct {
	Label  string
	Fields []FormField
}

// FormField is a single element of a form
type FormField struct {
	Label       string
	Description string
	Hint        string
	Type        string
}
