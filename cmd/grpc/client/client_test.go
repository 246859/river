package client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_Get(t *testing.T) {
	client, err := NewClient(context.Background(), Options{
		Target:   "localhost:6868",
		Password: "abcdefhijklmnopqrstuvwxyz",
	})
	assert.Nil(t, err)
	get, err := client.Get(context.Background(), []byte("a"))
	assert.Nil(t, err)
	t.Log(string(get))
}
