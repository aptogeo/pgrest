package main

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
	w.Write([]byte(restQuery.String()))
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
	{"/rest/User/1", "GET", "", &RestQuery{"User", "1", 0, 0, nil, nil}},
	{"/rest/User?offset=50&limit=10&sort=lastname,-firstname", "GET", "", &RestQuery{"User", "", 50, 10, nil, []*Sort{&Sort{"lastname", true}, &Sort{"firstname", false}}}},
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
		assert.Equal(t, string(body), mt.expected.String())
	}
}
