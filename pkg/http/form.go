package http

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flosch/pongo2/v4"
)

// ID computes the ID of a form by lower casing the name, then
// replacing all whitespace with underscores.
func (f Form) ID() string {
	return strings.ReplaceAll(strings.ToLower(f.Title), " ", "_")
}

// ID computes the ID of a field by lower casing hte name, then
// replacing all whitespace with underscores.
func (f FormField) ID() string {
	return strings.ReplaceAll(strings.ToLower(f.Label), " ", "_")
}

func (s *Server) loadForms() {
	files, _ := filepath.Glob("forms/*.json")

	for _, f := range files {
		formFile, err := os.Open(f)
		if err != nil {
			s.l.Warn("Error opening form", "file", f, "error", err)
			continue
		}
		bytes, err := io.ReadAll(formFile)
		if err != nil {
			s.l.Warn("Error reading form", "file", f, "error", err)
			continue
		}
		formFile.Close()
		formStruct := Form{}
		if err := json.Unmarshal(bytes, &formStruct); err != nil {
			s.l.Warn("Error loading form", "file", f, "error", err)
			continue
		}
		s.forms[formStruct.ID()] = formStruct
		s.l.Info("Loaded form", "form", formStruct.Title, "ID", formStruct.ID())
	}
}

func (s *Server) viewComplianceForm(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "view/form.p2", pongo2.Context{"form": s.forms["machine_compliance_check"]})
}
