package pgrest

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeserialize(t *testing.T) {
	config := NewConfig("/rest/", nil)
	config.AddResource(NewResource("Book", (*Book)(nil), All))
	engine := NewEngine(config)

	var err error
	var book *Book

	book = &Book{}
	err = engine.Deserialize(&RestQuery{Action: Post, Resource: "Book", ContentType: "application/json", Content: []byte("{\"Title\":\"a title\",\"NbPages\":520,\"UnknownField\":\"UnknownField\"}")}, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "a title")
	assert.Equal(t, book.NbPages, 520)

	book = &Book{}
	err = engine.Deserialize(&RestQuery{Action: Post, Resource: "Book", ContentType: "application/x-www-form-urlencoded", Content: []byte("Title=another title&NbPages=310&UnknownField=UnknownField")}, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "another title")
	assert.Equal(t, book.NbPages, 310)
}

func TestEngine(t *testing.T) {
	db, config := initTests(t)
	defer db.Close()
	engine := NewEngine(config)

	for _, author := range authors {
		content, err := json.Marshal(author)
		assert.Nil(t, err)
		res, err := engine.Execute(&RestQuery{Action: Post, Resource: "Author", ContentType: "application/json", Content: content})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resAuthor := res.(*Author)
		assert.NotEqual(t, resAuthor.ID, 0)
		assert.Equal(t, resAuthor.Firstname, author.Firstname)
		assert.Equal(t, resAuthor.Lastname, author.Lastname)
	}

	for _, book := range books {
		content, err := json.Marshal(book)
		assert.Nil(t, err)
		res, err := engine.Execute(&RestQuery{Action: Post, Resource: "Book", ContentType: "application/json", Content: content})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resBook := res.(*Book)
		assert.NotEqual(t, resBook.ID, 0)
		assert.NotEqual(t, resBook.AuthorID, 0)
		assert.Equal(t, resBook.Title, book.Title)
		assert.Equal(t, resBook.NbPages, 0)

		res, err = engine.Execute(&RestQuery{Action: Patch, Resource: "Book", Key: strconv.Itoa(resBook.ID), ContentType: "application/x-www-form-urlencoded", Content: []byte("NbPages=200")})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resBook = res.(*Book)
		assert.NotEqual(t, resBook.ID, 0)
		assert.NotEqual(t, resBook.AuthorID, 0)
		assert.Equal(t, resBook.Title, book.Title)
		assert.Equal(t, resBook.NbPages, 200)
	}

	var err error
	var res interface{}
	var page Page

	res, err = engine.Execute(&RestQuery{Action: Get, Resource: "Author"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*Page)
	assert.Equal(t, page.Count(), uint64(3))
	resAuthors := *page.Slice().(*[]Author)
	for _, author := range resAuthors {
		res, err = engine.Execute(&RestQuery{Action: Get, Resource: "Author", Key: strconv.Itoa(author.ID), Fields: []Field{Field{"*"}, Field{"Books"}}})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resAuthor := res.(*Author)
		assert.Equal(t, resAuthor.ID, author.ID)
		assert.Equal(t, author.Firstname, author.Firstname)
		assert.True(t, len(resAuthor.Books) > 0)
		for _, resBook := range resAuthor.Books {
			assert.NotNil(t, resBook.Title)
			assert.Equal(t, resBook.NbPages, 200)

			res, err = engine.Execute(&RestQuery{Action: Put, Resource: "Book", Key: strconv.Itoa(resBook.ID), ContentType: "application/x-www-form-urlencoded", Content: []byte("Title=" + resBook.Title + "_1&AuthorID=" + strconv.Itoa(resBook.AuthorID))})
			assert.Nil(t, err)
			assert.NotNil(t, res)
			resBook2 := res.(*Book)
			assert.NotEqual(t, resBook2.ID, 0)
			assert.NotEqual(t, resBook2.AuthorID, 0)
			assert.Equal(t, resBook2.Title, resBook.Title+"_1")
			assert.Equal(t, resBook2.NbPages, 0)
		}
	}

	_, err = engine.Execute(&RestQuery{Action: Delete, Resource: "Author", Key: "1"})
	res, err = engine.Execute(&RestQuery{Action: Get, Resource: "Author"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*Page)
	assert.Equal(t, page.Count(), uint64(2))

	_, err = engine.Execute(&RestQuery{Action: Delete, Resource: "Author", Key: "3"})
	res, err = engine.Execute(&RestQuery{Action: Get, Resource: "Author"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*Page)
	assert.Equal(t, page.Count(), uint64(1))
}
