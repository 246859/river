package client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClient_Get(t *testing.T) {
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	client, err := NewClient(timeout, Options{
		Target:   "localhost:6868",
		Password: "abcdefhijklmnopqrstuvwxyz",
	})
	assert.Nil(t, err)
	if err != nil {
		return
	}

	k, v := []byte("key"), []byte("value")
	ok, err := client.Put(context.Background(), k, v, -1)
	assert.Nil(t, err)
	assert.True(t, ok)

	get, err := client.Get(context.Background(), k)
	assert.Nil(t, err)
	assert.EqualValues(t, v, get)
	t.Log(string(get))
}
