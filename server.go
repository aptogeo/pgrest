package pgrest

import (
	"encoding/json"
	"net/http"
)

// Server structure
type Server struct {
	Engine
	next http.Handler
}

// NewServer constructs Server
func NewServer(config *Config) *Server {
	s := new(Server)
	s.config = config
	return s
}

// SetNextHandler sets next handler for middleware use
func (s *Server) SetNextHandler(next http.Handler) {
	s.next = next
}

// ServeHTTP serves rest request
func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	restQuery := RequestDecoder(request, s.Config())
	if restQuery != nil {
		res, err := s.Execute(restQuery)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonStr, err := json.Marshal(res)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if request.Method == "GET" {
			writer.WriteHeader(http.StatusOK)
		} else if request.Method == "POST" {
			writer.WriteHeader(http.StatusCreated)
		} else if request.Method == "PUT" {
			writer.WriteHeader(http.StatusOK)
		} else if request.Method == "PATCH" {
			writer.WriteHeader(http.StatusOK)
		} else if request.Method == "DELETE" {
			writer.WriteHeader(http.StatusNoContent)
		}
		writer.Write(jsonStr)
		return
	}
	if s.next != nil {
		s.next.ServeHTTP(writer, request)
	} else {
		http.Error(writer, "Request isn't rest request", http.StatusInternalServerError)
	}
}
