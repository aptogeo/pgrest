package pgrest

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func performRequest(h http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	restQuery := RestQueryFromRequest(r)
	if restQuery != nil {
		w.Write([]byte(restQuery.String()))
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var middlewareTests = []struct {
	uri      string
	method   string
	body     string
	expected *RestQuery
}{
	{"/rest/User/1", "GET", "", &RestQuery{Get, "User", "1", "", 0, 0, nil, nil}},
	{"/rest/User/1?fields=user.*,user.roles", "GET", "", &RestQuery{Get, "User", "1", "", 0, 0, []Field{Field{"user.*"}, Field{"user.roles"}}, nil}},
	{"/rest/User", "GET", "", &RestQuery{Get, "User", "", "", 0, 10, []Field{}, []Sort{}}},
	{"/rest/User?offset=50&limit=10&sort=lastname,-firstname", "GET", "", &RestQuery{Get, "User", "", "", 50, 10, []Field{}, []Sort{Sort{"lastname", true}, Sort{"firstname", false}}}},
	{"/rest/User?offset=60&limit=10&sort=lastname&fields=user.*,user.roles", "GET", "", &RestQuery{Get, "User", "", "", 60, 10, []Field{Field{"user.*"}, Field{"user.roles"}}, []Sort{Sort{"lastname", true}}}},
	{"/rest/User", "POST", "lastname=Doe&firstname=John", &RestQuery{Post, "User", "", "", 0, 0, nil, nil}},
	{"/rest/User/1", "PUT", "lastname=Doe&firstname=John", &RestQuery{Put, "User", "1", "", 0, 0, nil, nil}},
	{"/rest/User/1", "PATCH", "firstname=John", &RestQuery{Patch, "User", "1", "", 0, 0, nil, nil}},
	{"/rest/User/1", "DELETE", "", &RestQuery{Delete, "User", "1", "", 0, 0, nil, nil}},
	{"/rest", "GET", "", nil},
	{"/rest/Use/1", "POST", "", nil},
	{"/rest/User", "PUT", "", nil},
	{"/rest/User", "PATCH", "", nil},
	{"/rest/User", "DELETE", "", nil},
}

func TestDecodeRestQuery(t *testing.T) {
	handler := DecodeRestQuery(http.HandlerFunc(testHandler), "/rest/")

	ts := httptest.NewServer(handler)
	defer ts.Close()

	for _, mt := range middlewareTests {
		req, err := http.NewRequest(mt.method, ts.URL+mt.uri, bytes.NewBufferString(mt.body))
		check(err)
		res, err := http.DefaultClient.Do(req)
		check(err)
		body, err := ioutil.ReadAll(res.Body)
		check(err)
		res.Body.Close()
		check(err)
		if mt.expected != nil {
			assert.Equal(t, string(body), mt.expected.String())
		} else {
			assert.Equal(t, string(body), "")
		}
	}
}
