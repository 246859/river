package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadOption(t *testing.T) {
	_, err := ReadOption("./config.toml")
	assert.Nil(t, err)
}
