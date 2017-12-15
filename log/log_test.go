package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemovePortV4(t *testing.T) {

	assert := assert.New(t)
	actual := removePort("127.0.0.1:60000")
	assert.Equal("127.0.0.1", actual)
}

func TestRemovePortV6(t *testing.T) {

	assert := assert.New(t)
	actual := removePort("[::1]:60000")
	assert.Equal("[::1]", actual)
}
