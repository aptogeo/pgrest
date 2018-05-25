package pgrest

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// RequestDecoder decodes rest parameters from request
func RequestDecoder(request *http.Request, config *Config) *RestQuery {
	re := regexp.MustCompile("(" + config.Prefix() + ")(\\w+)/?(\\w+)?/?(\\w+)?")
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

		if offset, err := strconv.ParseUint(params.Get("offset"), 10, 64); err == nil {
			restQuery.Offset = offset
		}

		if limit, err := strconv.ParseUint(params.Get("limit"), 10, 64); err == nil {
			restQuery.Limit = limit
		}

		fieldsStr := strings.TrimSpace(params.Get("fields"))
		fieldsStrs := strings.Split(fieldsStr, ",")
		restQuery.Fields = make([]Field, 0)
		for _, s := range fieldsStrs {
			st := strings.TrimSpace(s)
			if st != "" {
				restQuery.Fields = append(restQuery.Fields, Field{st})
			}
		}

		sortStr := strings.TrimSpace(params.Get("sort"))
		sortStrs := strings.Split(sortStr, ",")
		restQuery.Sorts = make([]Sort, 0)
		for _, s := range sortStrs {
			st := strings.TrimSpace(s)
			if st != "" {
				if strings.HasPrefix(s, "-") {
					restQuery.Sorts = append(restQuery.Sorts, Sort{st[1:len(st)], false})
				} else {
					restQuery.Sorts = append(restQuery.Sorts, Sort{st, true})
				}
			}
		}

		return restQuery
	}
	return nil
}
