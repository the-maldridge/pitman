package http

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/flosch/pongo2/v4"
)

// filterGetValueByKey gives funcrtionality that really should have
// been in the template library to begin with and allows retrieving a
// single key from a map inside the template context.
func (s *Server) filterGetValueByKey(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	m, ok := in.Interface().(map[string]string)
	if !ok {
		s.l.Warn("Tried to convert something that isn't a map", "something", in)
		return pongo2.AsValue(""), nil
	}
	return pongo2.AsValue(m[param.String()]), nil
}

func (s *Server) filterGetValueAtIndex(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return in.Index(param.Integer()), nil
}

func (s *Server) csvToMap(reader io.Reader) []map[string]string {
	r := csv.NewReader(reader)
	rows := []map[string]string{}
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.l.Error("Error decoding CSV", "error", err)
			return nil
		}
		if header == nil {
			header = record
			for col := range header {
				header[col] = strings.ReplaceAll(header[col], "Team Name", "Name")
				header[col] = strings.ReplaceAll(header[col], "Team Number", "Number")
				header[col] = strings.ReplaceAll(header[col], "Hub Name", "Hub")
			}
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			rows = append(rows, dict)
		}
	}
	return rows
}
