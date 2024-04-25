package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) respondWithError(w http.ResponseWriter, status int, error string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"error":  error,
	})
}

func (s *Server) respondAny(w http.ResponseWriter, httpStatus int, obj any) {
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": httpStatus,
		"result": obj,
	})
}

func (s *Server) respondNoContent(w http.ResponseWriter, httpStatus int) {
	w.WriteHeader(httpStatus)
}
