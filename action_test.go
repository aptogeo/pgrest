package pgrest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAction(t *testing.T) {
	assert.NotEqual(t, None, Get)
	assert.NotEqual(t, None, Post)
	assert.NotEqual(t, None, Put)
	assert.NotEqual(t, None, Patch)
	assert.NotEqual(t, None, Delete)
	assert.NotEqual(t, None, All)
	assert.NotEqual(t, Get, Post)
	assert.NotEqual(t, Get, Put)
	assert.NotEqual(t, Get, Patch)
	assert.NotEqual(t, Get, Delete)
	assert.NotEqual(t, Get, All)
	assert.NotEqual(t, Post, Put)
	assert.NotEqual(t, Post, Patch)
	assert.NotEqual(t, Post, Delete)
	assert.NotEqual(t, Post, All)
	assert.NotEqual(t, Put, Patch)
	assert.NotEqual(t, Put, Delete)
	assert.NotEqual(t, Put, All)
	assert.NotEqual(t, Patch, Delete)
	assert.NotEqual(t, Patch, All)
	assert.NotEqual(t, Delete, All)
	assert.Equal(t, All, Get+Post+Put+Patch+Delete)
	assert.Equal(t, 0, int(None))
	assert.Equal(t, 1, int(Get))
	assert.Equal(t, 2, int(Post))
	assert.Equal(t, 4, int(Put))
	assert.Equal(t, 8, int(Patch))
	assert.Equal(t, 16, int(Delete))
	assert.Equal(t, 31, int(All))
}
