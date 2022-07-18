package http

import (
	"encoding/json"
	"net/http"
	"path"
	"sort"

	"github.com/flosch/pongo2/v4"
)

// This is a big ugly function that puts together the big board.  It
// gets all the teams and looks for a magic form called
// "master_status" to determine how far a team is in check in.
func (s *Server) viewBigBoard(w http.ResponseWriter, r *http.Request) {
	fname := "master_status"

	res := s.rdb.Keys(r.Context(), "teams/*")
	if res.Err() != nil {
		s.l.Warn("Error listing team IDs", "error", res.Err())
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": res.Err()})
		return
	}

	teams := make([]Team, len(res.Val()))
	for i, key := range res.Val() {
		tres := s.rdb.Get(r.Context(), key)
		bytes, err := tres.Bytes()
		if err != nil {
			s.l.Warn("Error retrieving team", "error", err, "key", key)
			s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err})
			return
		}
		team := Team{}
		if err := json.Unmarshal(bytes, &team); err != nil {
			s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err})
			return
		}
		teams[i] = team
	}
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Name < teams[j].Name
	})

	fields := []string{}
	for _, section := range s.forms[fname].Sections {
		for _, group := range section.Groups {
			for _, field := range group.Fields {
				if field.Type != "checkbox" {
					continue
				}
				fields = append(fields, section.ID()+"_"+group.ID()+"_"+field.ID())
			}
		}
	}

	fdata := []struct {
		Team   Team
		Status []string
		Done   bool
	}{}
	for _, team := range teams {
		res := s.rdb.Get(r.Context(), path.Join("forms", fname, team.Number))
		bytes, err := res.Bytes()
		if err != nil {
			s.l.Debug("Error retrieving form data", "error", err)
		}

		f := make(map[string]string)
		if err == nil {
			if err := json.Unmarshal(bytes, &f); err != nil {
				s.l.Warn("Error unmarshaling form data", "error", err)
			}
		}
		tfields := []string{}
		for k := range f {
			tfields = append(tfields, k)
		}
		fdata = append(fdata, struct {
			Team   Team
			Status []string
			Done bool
		}{
			Team:   team,
			Status: tfields,
			Done:   len(tfields) == len(fields),
		})
	}

	res2 := s.rdb.Get(r.Context(), path.Join("forms", "internal_configuration", "0"))
	bytes, err := res2.Bytes()
	if err != nil {
		s.l.Trace("No internal_configuration")
	}
	icfg := make(map[string]string)
	if err := json.Unmarshal(bytes, &icfg); err != nil {
		s.l.Warn("Error unmarshaling form data", "error", err)
	}

	ctx := pongo2.Context{
		"teams":  fdata,
		"fields": fields,
		"icfg":   icfg,
	}
	s.doTemplate(w, r, "view/big_board.p2", ctx)
}
