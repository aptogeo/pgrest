package pgrest_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/mathieumast/pgrest"
	"github.com/stretchr/testify/assert"
)

func TestDeserialize(t *testing.T) {
	config := pgrest.NewConfig("/rest/", nil)
	config.AddResource(pgrest.NewResource("Book", (*Book)(nil), pgrest.All))
	engine := pgrest.NewEngine(config)

	var err error
	var book *Book

	book = &Book{}
	err = engine.Deserialize(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/json", Content: []byte("{\"Title\":\"a title\",\"NbPages\":520,\"UnknownField\":\"UnknownField\"}")}, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "a title")
	assert.Equal(t, book.NbPages, 520)

	book = &Book{}
	err = engine.Deserialize(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/x-www-form-urlencoded", Content: []byte("Title=another title&NbPages=310&UnknownField=UnknownField")}, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "another title")
	assert.Equal(t, book.NbPages, 310)
}

func TestEngine(t *testing.T) {
	db, config := initTests(t)
	defer db.Close()
	engine := pgrest.NewEngine(config)

	var err error
	var content []byte
	var res interface{}
	var page pgrest.Page
	var resAuthor *Author
	var resAuthors []Author
	var resBook *Book

	for _, author := range authors {
		content, err = json.Marshal(author)
		assert.Nil(t, err)
		res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Author", ContentType: "application/json", Content: content})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resAuthor = res.(*Author)
		assert.NotEqual(t, resAuthor.ID, 0)
		assert.Equal(t, resAuthor.Firstname, author.Firstname)
		assert.Equal(t, resAuthor.Lastname, author.Lastname)
	}

	for _, book := range books {
		content, err = json.Marshal(book)
		assert.Nil(t, err)
		res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/json", Content: content})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resBook = res.(*Book)
		assert.NotEqual(t, resBook.ID, 0)
		assert.NotEqual(t, resBook.AuthorID, 0)
		assert.Equal(t, resBook.Title, book.Title)
		assert.Equal(t, resBook.NbPages, 0)

		res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Patch, Resource: "Book", Key: strconv.Itoa(resBook.ID), ContentType: "application/x-www-form-urlencoded", Content: []byte("NbPages=200")})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resBook = res.(*Book)
		assert.NotEqual(t, resBook.ID, 0)
		assert.NotEqual(t, resBook.AuthorID, 0)
		assert.Equal(t, resBook.Title, book.Title)
		assert.Equal(t, resBook.NbPages, 200)
	}

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*pgrest.Page)
	assert.Equal(t, page.Count, 3)
	resAuthors = *page.Slice.(*[]Author)
	assert.Equal(t, len(resAuthors), 3)
	for _, author := range resAuthors {
		res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author", Key: strconv.Itoa(author.ID), Fields: []*pgrest.Field{&pgrest.Field{Name: "*"}, &pgrest.Field{Name: "Books"}}})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resAuthor = res.(*Author)
		assert.Equal(t, resAuthor.ID, author.ID)
		assert.Equal(t, author.Firstname, author.Firstname)
		assert.True(t, len(resAuthor.Books) > 0)
		for _, resBook = range resAuthor.Books {
			assert.NotNil(t, resBook.Title)
			assert.Equal(t, resBook.NbPages, 200)

			res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Put, Resource: "Book", Key: strconv.Itoa(resBook.ID), ContentType: "application/x-www-form-urlencoded", Content: []byte("Title=" + resBook.Title + "_1&AuthorID=" + strconv.Itoa(resBook.AuthorID))})
			assert.Nil(t, err)
			assert.NotNil(t, res)
			resBook2 := res.(*Book)
			assert.NotEqual(t, resBook2.ID, 0)
			assert.NotEqual(t, resBook2.AuthorID, 0)
			assert.Equal(t, resBook2.Title, resBook.Title+"_1")
			assert.Equal(t, resBook2.NbPages, 0)
		}
	}

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Book", Filter: &pgrest.Filter{Op: pgrest.Ilk, Attr: "title", Value: "%lo%"}})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*pgrest.Page)
	assert.Equal(t, page.Count, 2)

	/*res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Book", Filter: &pgrest.Filter{Op: pgrest.Or, Filters: []*pgrest.Filter{&pgrest.Filter{Op: pgrest.Eq, Attr: "title", Value: "Le Petit Prince"}, &pgrest.Filter{Op: pgrest.Ilk, Attr: "title", Value: "NI"}}}})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*pgrest.Page)
	assert.Equal(t, page.Count, 3)*/

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author", Key: "12345"})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Delete, Resource: "Author", Key: "12345"})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Put, Resource: "Author", Key: "12345", ContentType: "application/x-www-form-urlencoded", Content: []byte("Firstname=Firstname&Lastanme=Lastname")})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Patch, Resource: "Author", Key: "12345", ContentType: "application/x-www-form-urlencoded", Content: []byte("Firstname=Firstname")})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	_, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Delete, Resource: "Author", Key: "1"})
	assert.Nil(t, err)
	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*pgrest.Page)
	assert.Equal(t, page.Count, 2)
	resAuthors = *page.Slice.(*[]Author)
	assert.Equal(t, len(resAuthors), 2)

	_, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Delete, Resource: "Author", Key: "3"})
	assert.Nil(t, err)
	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*pgrest.Page)
	assert.Equal(t, page.Count, 1)
	resAuthors = *page.Slice.(*[]Author)
	assert.Equal(t, len(resAuthors), 1)
}
