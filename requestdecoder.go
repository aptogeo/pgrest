package pgrest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// RequestDecoder decodes rest parameters from request
func RequestDecoder(request *http.Request, config *Config) *RestQuery {
	re := regexp.MustCompile("(" + config.Prefix() + ")([^/\\?]+)/?([^/\\?]+)?/?([^/\\?]+)?")
	res := re.FindStringSubmatch(request.RequestURI)
	action := None
	if request.Method == "GET" {
		action = Get
	} else if request.Method == "POST" {
		action = Post
	} else if request.Method == "PUT" {
		action = Put
	} else if request.Method == "PATCH" {
		action = Patch
	} else if request.Method == "DELETE" {
		action = Delete
	}
	if res != nil && res[4] == "" && action != None {
		restQuery := &RestQuery{Action: action, Offset: 0, Limit: 10}
		restQuery.Resource = res[2]
		restQuery.Key = res[3]

		params := request.URL.Query()

		restQuery.Content, _ = ioutil.ReadAll(request.Body)

		restQuery.ContentType = request.Header.Get("Content-Type")
		if restQuery.ContentType == "" {
			restQuery.ContentType = config.DefaultContentType()
		}

		restQuery.Accept = request.Header.Get("Accept")
		if restQuery.Accept == "" {
			restQuery.Accept = config.DefaultAccept()
		}

		if offset, err := strconv.ParseInt(params.Get("offset"), 10, 64); err == nil {
			restQuery.Offset = int(offset)
		}

		if limit, err := strconv.ParseInt(params.Get("limit"), 10, 64); err == nil {
			restQuery.Limit = int(limit)
		}

		fieldsStr := strings.TrimSpace(params.Get("fields"))
		fieldsStrs := strings.Split(fieldsStr, ",")
		restQuery.Fields = make([]*Field, 0)
		for _, s := range fieldsStrs {
			st := strings.TrimSpace(s)
			if st != "" {
				restQuery.Fields = append(restQuery.Fields, &Field{st})
			}
		}

		sortStr := strings.TrimSpace(params.Get("sort"))
		sortStrs := strings.Split(sortStr, ",")
		restQuery.Sorts = make([]*Sort, 0)
		for _, s := range sortStrs {
			st := strings.TrimSpace(s)
			if st != "" {
				if strings.HasPrefix(s, "-") {
					restQuery.Sorts = append(restQuery.Sorts, &Sort{st[1:len(st)], false})
				} else {
					restQuery.Sorts = append(restQuery.Sorts, &Sort{st, true})
				}
			}
		}

		filterStr := strings.TrimSpace(params.Get("filter"))
		restQuery.Filter = &Filter{}
		if strings.HasPrefix(filterStr, "{") {
			json.Unmarshal([]byte(filterStr), restQuery.Filter)
		}

		if debug, err := strconv.ParseBool(params.Get("debug")); err == nil {
			restQuery.Debug = debug
		}

		return restQuery.WithContext(request.Context())
	}
	return nil
}
