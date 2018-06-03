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
			if cerr, ok := err.(*Error); ok {
				http.Error(writer, cerr.Error(), cerr.StatusCode())
			}
		} else if res == nil {
			http.Error(writer, "resource not found", http.StatusNotFound)
		} else {
			jsonStr, err := json.Marshal(res)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			} else {
				if restQuery.Action == Get {
					writer.WriteHeader(http.StatusOK)
				} else if restQuery.Action == Post {
					writer.WriteHeader(http.StatusCreated)
				} else if restQuery.Action == Put {
					writer.WriteHeader(http.StatusOK)
				} else if restQuery.Action == Patch {
					writer.WriteHeader(http.StatusOK)
				} else if restQuery.Action == Delete {
					writer.WriteHeader(http.StatusNoContent)
				} else {
					writer.WriteHeader(http.StatusOK)
				}
				writer.Write(jsonStr)
			}
		}
	} else {
		if s.next != nil {
			s.next.ServeHTTP(writer, request)
		} else {
			http.Error(writer, "Request isn't rest request", http.StatusInternalServerError)
		}
	}
}
