package http

import (
	"net/http"
)

func (s *Server) viewTeamStatus(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "view/team_status.p2", nil)
}
