package pgrest_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aptogeo/pgrest"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	db, config := initTests(t)
	defer db.Close()
	server := pgrest.NewServer(config)

	ts := httptest.NewServer(server)
	defer ts.Close()

	var err error
	var req *http.Request
	var body []byte
	var res *http.Response
	var page *pgrest.Page
	var resAuthor *Author
	var resBook *Book

	for _, author := range authors {
		content, err := json.Marshal(author)
		assert.Nil(t, err)
		req, err = http.NewRequest("POST", ts.URL+"/rest/Author", bytes.NewBuffer(content))
		assert.Nil(t, err)
		res, err = http.DefaultClient.Do(req)
		assert.Nil(t, err)
		body, err = ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		err = res.Body.Close()
		assert.Nil(t, err)
		resAuthor = &Author{}
		err = json.Unmarshal(body, resAuthor)
		assert.Nil(t, err)
		assert.NotEqual(t, resAuthor.ID, 0)
		assert.Equal(t, resAuthor.Firstname, author.Firstname)
		assert.Equal(t, resAuthor.Lastname, author.Lastname)
	}

	req, err = http.NewRequest("GET", ts.URL+"/rest/Author", bytes.NewBufferString(""))
	assert.Nil(t, err)
	res, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = res.Body.Close()
	assert.Nil(t, err)
	page = &pgrest.Page{}
	err = json.Unmarshal(body, page)
	assert.Nil(t, err)
	assert.NotNil(t, page)
	assert.Equal(t, page.Count, 3)

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
		resBook = &Book{}
		err = json.Unmarshal(body, resBook)
		assert.Nil(t, err)
		assert.NotEqual(t, resBook.ID, 0)
		assert.NotEqual(t, resBook.AuthorID, 0)
		assert.Equal(t, resBook.Title, book.Title)
		assert.Equal(t, resBook.NbPages, 0)
	}

	req, err = http.NewRequest("GET", ts.URL+"/rest/Book", bytes.NewBufferString(""))
	assert.Nil(t, err)
	res, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = res.Body.Close()
	assert.Nil(t, err)
	page = &pgrest.Page{}
	err = json.Unmarshal(body, page)
	assert.Nil(t, err)
	assert.NotNil(t, page)
	assert.Equal(t, page.Count, 12)

	req, err = http.NewRequest("GET", ts.URL+"/rest/Book", bytes.NewBufferString(""))
	assert.Nil(t, err)
	req.Header.Set("Accept", "application/x-msgpack")
	res, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	err = res.Body.Close()
	assert.Nil(t, err)
}
