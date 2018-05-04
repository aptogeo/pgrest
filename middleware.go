package pgrest

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// RestQueryKey context key for the rest query.
const restQueryKey key = 658

// DecodeRestQuery decodes rest parameters
func DecodeRestQuery(next http.Handler, pattern string) http.HandlerFunc {
	// log
	var log = log.New(os.Stdout, "RestQuery "+pattern+":", log.Lshortfile)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var restQuery *RestQuery
		var resource, key string
		re := regexp.MustCompile(pattern + "(\\w+)(/(\\w*))?")
		res := re.FindStringSubmatch(r.RequestURI)
		if len(res) > 1 {
			resource = res[1]
		}
		if len(res) > 3 {
			key = res[3]
		}

		var err error
		var offset, limit uint64
		var body string

		if resource != "" {
			if key != "" {
				if r.Method == "GET" {
					restQuery = &RestQuery{Get, resource, key, "", 0, 0, nil, nil}
				} else if r.Method == "PUT" {
					if bytes, err := ioutil.ReadAll(r.Body); err == nil {
						body = string(bytes)
					}
					restQuery = &RestQuery{Put, resource, key, body, 0, 0, nil, nil}
				} else if r.Method == "PATCH" {
					if bytes, err := ioutil.ReadAll(r.Body); err == nil {
						body = string(bytes)
					}
					restQuery = &RestQuery{Patch, resource, key, body, 0, 0, nil, nil}
				} else if r.Method == "DELETE" {
					restQuery = &RestQuery{Delete, resource, key, "", 0, 0, nil, nil}
				}
			} else {
				if r.Method == "GET" {
					params := r.URL.Query()

					offsetStr := params.Get("offset")
					if offset, err = strconv.ParseUint(offsetStr, 10, 64); err != nil {
						offset = 0
					}

					limitStr := params.Get("limit")
					if limit, err = strconv.ParseUint(limitStr, 10, 64); err != nil {
						limit = 10
					}

					var fields []*Field
					fieldsStr := strings.TrimSpace(params.Get("fields"))
					fieldsStrs := strings.Split(fieldsStr, ",")
					fields = make([]*Field, 0)
					for _, s := range fieldsStrs {
						st := strings.TrimSpace(s)
						if st != "" {
							fields = append(fields, &Field{st})
						}
					}

					var sorts []*Sort
					sortStr := strings.TrimSpace(params.Get("sort"))
					sortStrs := strings.Split(sortStr, ",")
					sorts = make([]*Sort, 0)
					for _, s := range sortStrs {
						st := strings.TrimSpace(s)
						if st != "" {
							if strings.HasPrefix(s, "-") {
								sorts = append(sorts, &Sort{st[1:len(st)], false})
							} else {
								sorts = append(sorts, &Sort{st, true})
							}
						}
					}
					restQuery = &RestQuery{Get, resource, "", "", offset, limit, fields, sorts}
				} else if r.Method == "POST" {
					if bytes, err := ioutil.ReadAll(r.Body); err == nil {
						body = string(bytes)
					}
					restQuery = &RestQuery{Post, resource, "", body, 0, 0, nil, nil}
				}
			}
		}
		if restQuery != nil {
			log.Println(restQuery)
			r = r.WithContext(context.WithValue(r.Context(), restQueryKey, restQuery))
			log.Println(r.Context())
		}
		next.ServeHTTP(w, r)
	})
}

// RestQueryFromRequest can be used to obtain the RestQuery from the request.
func RestQueryFromRequest(r *http.Request) *RestQuery {
	obj := r.Context().Value(restQueryKey)
	if obj == nil {
		return nil
	}
	return obj.(*RestQuery)
}

// RestQueryFromContext can be used to obtain the RestQuery from the context.
func RestQueryFromContext(ctx context.Context) *RestQuery {
	obj := ctx.Value(restQueryKey)
	if obj == nil {
		return nil
	}
	return obj.(*RestQuery)
}

// Handle with RestQuery
func Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, DecodeRestQuery(handler, pattern))
}

// HandleFunc with RestQuery
func HandleFunc(pattern string, handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	Handle(pattern, http.HandlerFunc(handlerFunc))
}
