package pgrest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func nothingToDOHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Nothing to do"))
}

func TestServer(t *testing.T) {
	db, config := initTests(t)
	defer db.Close()
	server := NewServer(config)

	ts := httptest.NewServer(server)
	defer ts.Close()

	for _, author := range authors {
		content, err := json.Marshal(author)
		assert.Nil(t, err)
		req, err := http.NewRequest("POST", ts.URL+"/rest/Author", bytes.NewBuffer(content))
		assert.Nil(t, err)
		res, err := http.DefaultClient.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		err = res.Body.Close()
		assert.Nil(t, err)
		resAuthor := &Author{}
		err = json.Unmarshal(body, resAuthor)
		assert.Nil(t, err)
		assert.NotEqual(t, resAuthor.ID, 0)
		assert.Equal(t, resAuthor.Firstname, author.Firstname)
		assert.Equal(t, resAuthor.Lastname, author.Lastname)
	}

	for _, book := range books {
		content, err := json.Marshal(book)
		assert.Nil(t, err)
		req, err := http.NewRequest("POST", ts.URL+"/rest/Book", bytes.NewBuffer(content))
		assert.Nil(t, err)
		res, err := http.DefaultClient.Do(req)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		err = res.Body.Close()
		assert.Nil(t, err)
		resBook := &Book{}
		err = json.Unmarshal(body, resBook)
		assert.Nil(t, err)
		assert.NotEqual(t, resBook.ID, 0)
		assert.NotEqual(t, resBook.AuthorID, 0)
		assert.Equal(t, resBook.Title, book.Title)
		assert.Equal(t, resBook.NbPages, 0)
	}
}
