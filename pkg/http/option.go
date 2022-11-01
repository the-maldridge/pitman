package http

import (
	"github.com/hashicorp/go-hclog"
)

// Option sets parameters of the server.
type Option func(*Server)

// WithLogger sets the logger on the server.
func WithLogger(l hclog.Logger) Option {
	return func(s *Server) { s.l = l.Named("http") }
}

// WithStorage sets the storage engine for the server.
func WithStorage(kv KV) Option {
	return func(s *Server) { s.kv = kv }
}
