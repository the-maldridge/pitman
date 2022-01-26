package http

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
)

// ID computes the ID of a form by lower casing the name, then
// replacing all whitespace with underscores.
func (f Form) ID() string {
	return strings.ReplaceAll(strings.ToLower(f.Title), " ", "_")
}

// ID computes the ID by lower casing the name, then replacing all
// whitespace with underscores.
func (f FormSection) ID() string {
	return strings.ReplaceAll(strings.ToLower(f.Label), " ", "_")
}

// ID computes the ID by lower casing the name, then replacing all
// whitespace with underscores.
func (f FormGroup) ID() string {
	return strings.ReplaceAll(strings.ToLower(f.Label), " ", "_")
}

// ID computes the ID of a field by lower casing the name, then
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

func (s *Server) viewForm(w http.ResponseWriter, r *http.Request) {
	// team number 0 is special and is used for internal forms
	// that configure the system.  This is used to suppress some
	// errors below where the system reaches for a team and it
	// doesn't actually exist.  This is pretty safe since team
	// number zero can't actually be issued for other reasons.
	teamnum := chi.URLParam(r, "id")
	fname := chi.URLParam(r, "form")

	tres := s.rdb.Get(r.Context(), path.Join("teams", teamnum))
	bytes, err := tres.Bytes()
	if err != nil && teamnum != "0" {
		s.l.Warn("Error retrieving team", "error", err, "key", path.Join("teams", teamnum))
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err})
		return
	}
	team := Team{}
	if err := json.Unmarshal(bytes, &team); err != nil && teamnum != "0" {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err})
		return
	}

	res := s.rdb.Get(r.Context(), path.Join("forms", fname, teamnum))
	bytes, err = res.Bytes()
	if err != nil {
		s.l.Debug("Error retrieving form data", "error", err)
	}

	fdata := make(map[string]string)
	if err := json.Unmarshal(bytes, &fdata); err != nil {
		s.l.Warn("Error unmarshaling form data", "error", err)
	}

	ctx := pongo2.Context{
		"team": team,
		"form":  s.forms[fname],
		"fdata": fdata,
	}

	s.doTemplate(w, r, "view/form.p2", ctx)
}

func (s *Server) submitForm(w http.ResponseWriter, r *http.Request) {
	teamnum := chi.URLParam(r, "id")
	fname := chi.URLParam(r, "form")
	r.ParseForm()

	fdata := make(map[string]string)
	for key := range r.Form {
		fdata[key] = r.FormValue(key)
	}

	bytes, err := json.Marshal(fdata)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	if status := s.rdb.Set(r.Context(), path.Join("forms", fname, teamnum), bytes, 0); status.Err() != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": status.Err()})
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
