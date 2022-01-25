package http

import (
	"github.com/flosch/pongo2/v4"
)

// filterGetValueByKey gives funcrtionality that really should have
// been in the template library to begin with and allows retrieving a
// single key from a map inside the template context.
func (s *Server) filterGetValueByKey(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	m := in.Interface().(map[string]string)
	return pongo2.AsValue(m[param.String()]), nil
}
