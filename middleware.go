package main

import (
	"context"
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

// RestQueryKey Context key for the rest query.
const restQueryKey key = 658

// DecodeRestQuery decode rest parameters
func DecodeRestQuery(next http.Handler, pattern string) http.HandlerFunc {
	// log
	var log = log.New(os.Stdout, "RestQuery "+pattern+":", log.Lshortfile)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			var err error
			var resource, key string
			var offset, limit uint64
			var restQuery *RestQuery

			re := regexp.MustCompile(pattern + "(\\w+)(/(\\w*))?")
			res := re.FindStringSubmatch(r.RequestURI)

			if len(res) > 1 {
				resource = res[1]
			}
			if len(res) > 3 {
				key = res[3]
			}

			params := r.URL.Query()

			if resource == "" {
				restQuery = &RestQuery{"", "", 0, 0, nil, nil}
			} else if key != "" {
				restQuery = &RestQuery{resource, key, 0, 0, nil, nil}
			} else {
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
				if fieldsStr != "" {
					fieldsStrs := strings.Split(fieldsStr, ",")
					fields = make([]*Field, len(fieldsStrs))
					for i, s := range fieldsStrs {
						st := strings.TrimSpace(s)
						if st != "" {
							fields[i] = &Field{st}
						}
					}
				}

				var sorts []*Sort
				sortStr := strings.TrimSpace(params.Get("sort"))
				sortStrs := strings.Split(sortStr, ",")
				sorts = make([]*Sort, len(sortStrs))
				for i, s := range sortStrs {
					st := strings.TrimSpace(s)
					if st != "" {
						if strings.HasPrefix(s, "-") {
							sorts[i] = &Sort{st[1:len(st)], false}
						} else {
							sorts[i] = &Sort{st, true}
						}
					}
				}
				restQuery = &RestQuery{resource, key, offset, limit, fields, sorts}
			}

			log.Println(restQuery)
			r = r.WithContext(context.WithValue(r.Context(), restQueryKey, restQuery))
			log.Println(r.Context())
		}
		next.ServeHTTP(w, r)
	})
}

// RestQueryFromRequest can be used to obtain the RestQuery from the request.
func RestQueryFromRequest(r *http.Request) *RestQuery {
	return r.Context().Value(restQueryKey).(*RestQuery)
}

// RestQueryFromContext can be used to obtain the RestQuery from the context.
func RestQueryFromContext(ctx context.Context) *RestQuery {
	return ctx.Value(restQueryKey).(*RestQuery)
}

// Handle with RestQuery
func Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, DecodeRestQuery(handler, pattern))
}

// HandleFunc with RestQuery
func HandleFunc(pattern string, handlerFunc func(w http.ResponseWriter, r *http.Request)) {
	Handle(pattern, http.HandlerFunc(handlerFunc))
}
