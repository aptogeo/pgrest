package pgrest_test

import (
	"testing"

	"github.com/aptogeo/pgrest"
	"github.com/stretchr/testify/assert"
)

func TestAction(t *testing.T) {
	assert.NotEqual(t, pgrest.None, pgrest.Get)
	assert.NotEqual(t, pgrest.None, pgrest.Post)
	assert.NotEqual(t, pgrest.None, pgrest.Put)
	assert.NotEqual(t, pgrest.None, pgrest.Patch)
	assert.NotEqual(t, pgrest.None, pgrest.Delete)
	assert.NotEqual(t, pgrest.None, pgrest.All)
	assert.NotEqual(t, pgrest.Get, pgrest.Post)
	assert.NotEqual(t, pgrest.Get, pgrest.Put)
	assert.NotEqual(t, pgrest.Get, pgrest.Patch)
	assert.NotEqual(t, pgrest.Get, pgrest.Delete)
	assert.NotEqual(t, pgrest.Get, pgrest.All)
	assert.NotEqual(t, pgrest.Post, pgrest.Put)
	assert.NotEqual(t, pgrest.Post, pgrest.Patch)
	assert.NotEqual(t, pgrest.Post, pgrest.Delete)
	assert.NotEqual(t, pgrest.Post, pgrest.All)
	assert.NotEqual(t, pgrest.Put, pgrest.Patch)
	assert.NotEqual(t, pgrest.Put, pgrest.Delete)
	assert.NotEqual(t, pgrest.Put, pgrest.All)
	assert.NotEqual(t, pgrest.Patch, pgrest.Delete)
	assert.NotEqual(t, pgrest.Patch, pgrest.All)
	assert.NotEqual(t, pgrest.Delete, pgrest.All)
	assert.Equal(t, pgrest.All, pgrest.Get+pgrest.Post+pgrest.Put+pgrest.Patch+pgrest.Delete)
	assert.Equal(t, 0, int(pgrest.None))
	assert.Equal(t, 1, int(pgrest.Get))
	assert.Equal(t, 2, int(pgrest.Post))
	assert.Equal(t, 4, int(pgrest.Put))
	assert.Equal(t, 8, int(pgrest.Patch))
	assert.Equal(t, 16, int(pgrest.Delete))
	assert.Equal(t, 31, int(pgrest.All))
}
