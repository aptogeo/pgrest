package pgrest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func decodeHandler(prefix string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restQuery := RequestDecoder(r, NewConfig("/rest/", nil))
		if restQuery != nil {
			w.Write([]byte(restQuery.String()))
		}
	})
}

var requestDecoderTests = []struct {
	uri      string
	method   string
	expected *RestQuery
}{
	{"/rest/User/1", "GET", &RestQuery{Action: Get, Resource: "User", Key: "1"}},
	{"/rest/User/1?fields=*,Roles", "GET", &RestQuery{Action: Get, Resource: "User", Key: "1", Offset: 0, Limit: 10, Fields: []Field{Field{"*"}, Field{"Roles"}}}},
	{"/rest/User", "GET", &RestQuery{Action: Get, Resource: "User", Offset: 0, Limit: 10, Fields: []Field{}, Sorts: []Sort{}}},
	{"/rest/User?offset=50&limit=10&sort=lastname,-firstname", "GET", &RestQuery{Action: Get, Resource: "User", Offset: 50, Limit: 10, Fields: []Field{}, Sorts: []Sort{Sort{"lastname", true}, Sort{"firstname", false}}}},
	{"/rest/User?offset=60&limit=10&sort=lastname&fields=user.*,user.roles", "GET", &RestQuery{Action: Get, Resource: "User", Offset: 60, Limit: 10, Fields: []Field{Field{"user.*"}, Field{"user.roles"}}, Sorts: []Sort{Sort{"lastname", true}}}},
	{"/rest/User", "POST", &RestQuery{Action: Post, Resource: "User"}},
	{"/rest/User/1", "PUT", &RestQuery{Action: Put, Resource: "User", Key: "1"}},
	{"/rest/User/1", "PATCH", &RestQuery{Action: Patch, Resource: "User", Key: "1"}},
	{"/rest/User/1", "DELETE", &RestQuery{Action: Delete, Resource: "User", Key: "1"}},
	{"/rest/User/specific/otherservice", "GET", nil},
	{"/rest", "GET", nil},
	{"/", "GET", nil},
}

func TestRequestDecoder(t *testing.T) {
	ts := httptest.NewServer(decodeHandler("/rest/"))
	defer ts.Close()

	for _, rt := range requestDecoderTests {
		req, err := http.NewRequest(rt.method, ts.URL+rt.uri, bytes.NewBufferString(""))
		assert.Nil(t, err)
		res, err := http.DefaultClient.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		err = res.Body.Close()
		assert.Nil(t, err)
		if rt.expected != nil {
			assert.Equal(t, string(body), rt.expected.String())
		} else {
			assert.Equal(t, string(body), "")
		}
	}
}
