package http

import (
	"encoding/json"
	"net/http"
	"path"
	"sort"

	"github.com/flosch/pongo2/v4"
)

func (s *Server) viewTeamStatus(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "view/team_status.p2", nil)
}

func (s *Server) viewTeams(w http.ResponseWriter, r *http.Request) {
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

	ctx := pongo2.Context{
		"teams": teams,
		"forms": s.forms,
	}

	s.doTemplate(w, r, "view/team_admin.p2", ctx)
}

func (s *Server) submitTeams(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	f, _, err := r.FormFile("teams_file")
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	defer f.Close()
	teams := s.csvToMap(f)

	for _, team := range teams {
		teamBytes, err := json.Marshal(team)
		if err != nil {
			s.l.Warn("Error marshaling team", "team", team["Number"], "error", err)
			continue
		}

		if res := s.rdb.Set(r.Context(), path.Join("teams", team["Number"]), teamBytes, 0); res.Err() != nil {
			s.l.Warn("Error Loading Team", "team", team["Number"], "error", res.Err())
		}
	}
	http.Redirect(w, r, "/admin/teams", http.StatusSeeOther)
}
