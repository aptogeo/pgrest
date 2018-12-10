package pgrest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/vmihailenco/msgpack"
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
			serialized, contentType, err := s.Serialize(restQuery, res)
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
				writer.Header().Add("Content-Type", contentType+"; charset=utf-8")
				writer.Write(serialized)
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

// Serialize serializes data into entity
func (s *Server) Serialize(restQuery *RestQuery, entity interface{}) ([]byte, string, error) {
	if regexp.MustCompile("[+-/]json($|[+-;])").MatchString(restQuery.Accept) {
		data, err := json.Marshal(entity)
		return data, "application/json", err
	} else if regexp.MustCompile("[+-/]msgpack($|[+-;])").MatchString(restQuery.Accept) {
		data, err := msgpack.Marshal(entity)
		return data, "application/msgpack", err
	}
	return nil, "plain/text", NewErrorBadRequest(fmt.Sprintf("Unknown accept '%v'", restQuery.Accept))
}
