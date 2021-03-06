package pgrest_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/aptogeo/pgrest"
	"github.com/stretchr/testify/assert"
)

func TestDeserialize(t *testing.T) {
	config := pgrest.NewConfig("/rest/", nil)
	resource := pgrest.NewResource("Book", (*Book)(nil), pgrest.All)
	config.AddResource(resource)
	engine := pgrest.NewEngine(config)

	var err error
	var content []byte
	var book *Book

	book = &Book{}
	err = engine.Deserialize(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/json", Content: []byte("{\"Title\":\"a title\",\"NbPages\":520,\"UnknownField\":\"UnknownField\"}")}, resource, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "a title")
	assert.Equal(t, book.NbPages, 520)
	err = engine.Deserialize(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/json", Content: []byte("{\"NbPages\":230}")}, resource, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "a title")
	assert.Equal(t, book.NbPages, 230)

	book = &Book{}
	err = engine.Deserialize(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/x-www-form-urlencoded", Content: []byte("Title=another title&NbPages=310&UnknownField=UnknownField")}, resource, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "another title")
	assert.Equal(t, book.NbPages, 310)
	err = engine.Deserialize(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/x-www-form-urlencoded", Content: []byte("NbPages=450")}, resource, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "another title")
	assert.Equal(t, book.NbPages, 450)

	content, err = msgpack.Marshal(&Book{Title: "msgpack title", NbPages: 480})
	assert.Nil(t, err)
	book = &Book{}
	err = engine.Deserialize(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/x-msgpack", Content: content}, resource, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "msgpack title")
	assert.Equal(t, book.NbPages, 480)
	content, err = msgpack.Marshal(&PageOnly{NbPages: 110})
	assert.Nil(t, err)
	err = engine.Deserialize(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Book", ContentType: "application/x-msgpack", Content: content}, resource, book)
	assert.Nil(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, book.Title, "msgpack title")
	assert.Equal(t, book.NbPages, 110)
}

func TestPostPatchGetDelete(t *testing.T) {
	db, config := initTests(t)
	defer db.Close()
	engine := pgrest.NewEngine(config)

	var err error
	var content []byte
	var res interface{}
	var page pgrest.Page
	var resTodo *Todo
	var resAuthor *Author
	var resAuthors []Author
	var resBook *Book

	for _, todo := range todos {
		content, err = json.Marshal(todo)
		assert.Nil(t, err)
		res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Todo", ContentType: "application/json", Content: content})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resTodo = res.(*Todo)
		assert.NotEqual(t, resTodo.ID, uuid.Nil)
		assert.NotEqual(t, resTodo.Text, "")
	}

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
		res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author", Key: strconv.Itoa(author.ID), Fields: []*pgrest.Field{{Name: "*"}}, Relations: []*pgrest.Relation{{Name: "Books"}}})
		assert.Nil(t, err)
		assert.NotNil(t, res)
		resAuthor = res.(*Author)
		assert.Equal(t, resAuthor.ID, author.ID)
		assert.Equal(t, resAuthor.Firstname, author.Firstname)
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

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author", Key: "1"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	resAuthor = res.(*Author)
	assert.Equal(t, resAuthor.Lastname, "de Saint Exupéry")
	assert.Equal(t, resAuthor.TransientField, "\"$user\", public")

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author", Key: "1", SearchPath: "public, other"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	resAuthor = res.(*Author)
	assert.Equal(t, resAuthor.Lastname, "de Saint Exupéry")
	assert.Equal(t, resAuthor.TransientField, "public, other")

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Book", Filter: &pgrest.Filter{Op: pgrest.Or, Filters: []*pgrest.Filter{{Op: pgrest.Ilk, Attr: "title", Value: "%lo%"}, {Op: pgrest.Ilk, Attr: "title", Value: "%ta%"}}}})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*pgrest.Page)
	assert.Equal(t, page.Count, 4)

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author", Filter: &pgrest.Filter{Op: pgrest.In, Attr: "firstname", Value: []string{"Antoine", "Franz"}}})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*pgrest.Page)
	assert.Equal(t, page.Count, 2)

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Get, Resource: "Author", Filter: &pgrest.Filter{Op: pgrest.In, Attr: "firstname", Value: []string{"Antoine", "Franz"}}})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	page = *res.(*pgrest.Page)
	assert.Equal(t, page.Count, 2)

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

func TestFormUrlencoded(t *testing.T) {
	db, config := initTests(t)
	defer db.Close()
	engine := pgrest.NewEngine(config)

	var err error
	var res interface{}
	var resAuthor *Author

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Author", ContentType: "application/x-www-form-urlencoded", Content: []byte("Firstname=Firstname&Lastname=Lastname")})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	resAuthor = res.(*Author)
	assert.NotEqual(t, resAuthor.ID, 0)
	assert.Equal(t, resAuthor.Firstname, "Firstname")
	assert.Equal(t, resAuthor.Lastname, "Lastname")
}

func TestMsgpack(t *testing.T) {
	db, config := initTests(t)
	defer db.Close()
	engine := pgrest.NewEngine(config)

	var err error
	var content []byte
	var res interface{}
	var resAuthor *Author

	resAuthor = &Author{Firstname: "MsgpackFirstname", Lastname: "MsgpackLastname", Picture: []byte{187, 163, 35, 30}}
	content, err = msgpack.Marshal(resAuthor)
	assert.Nil(t, err)

	res, err = engine.Execute(&pgrest.RestQuery{Action: pgrest.Post, Resource: "Author", ContentType: "application/x-msgpack", Content: content})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	resAuthor = res.(*Author)
	assert.NotEqual(t, resAuthor.ID, 0)
	assert.Equal(t, resAuthor.Firstname, "MsgpackFirstname")
	assert.Equal(t, resAuthor.Lastname, "MsgpackLastname")
	assert.Equal(t, resAuthor.Picture, []byte{187, 163, 35, 30})
}

func TestSearchPath(t *testing.T) {
	db, _ := initTests(t)
	defer db.Close()

	var searchPathTx1 string
	var searchPathTx2 string
	var searchPathDB string

	tx1, err := db.Begin()
	assert.Nil(t, err)

	_, err = db.QueryOne(pg.Scan(&searchPathDB), "SHOW search_path")
	assert.Nil(t, err)
	assert.Equal(t, "\"$user\", public", searchPathDB)

	tx1.Exec("SET search_path = test1")

	_, err = tx1.QueryOne(pg.Scan(&searchPathTx1), "SHOW search_path")
	assert.Nil(t, err)
	assert.Equal(t, "test1", searchPathTx1)

	_, err = db.QueryOne(pg.Scan(&searchPathDB), "SHOW search_path")
	assert.Nil(t, err)
	assert.Equal(t, "\"$user\", public", searchPathDB)

	tx2, err := db.Begin()
	assert.Nil(t, err)

	tx2.Exec("SET search_path = test2")

	_, err = tx2.QueryOne(pg.Scan(&searchPathTx2), "SHOW search_path")
	assert.Nil(t, err)
	assert.Equal(t, "test2", searchPathTx2)

	_, err = db.QueryOne(pg.Scan(&searchPathDB), "SHOW search_path")
	assert.Nil(t, err)
	assert.Equal(t, "\"$user\", public", searchPathDB)

	err = tx1.Rollback()
	assert.Nil(t, err)

	err = tx2.Rollback()
	assert.Nil(t, err)

	_, err = db.QueryOne(pg.Scan(&searchPathDB), "SHOW search_path")
	assert.Nil(t, err)
	assert.Equal(t, "\"$user\", public", searchPathDB)
}
