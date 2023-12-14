package client

import (
	"context"
	"fmt"
	riverdb "github.com/246859/river"
	"github.com/246859/river/cmd/river/riverpb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClient_Timeout(t *testing.T) {
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	_, err := NewClient(timeout, Options{
		Target:   "unknown",
		Password: "abcdefhijklmnopqrstuvwxyz",
	})
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

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
	defer client.Close()

	k, v := []byte("key"), []byte("value")
	ok, err := client.Put(context.Background(), k, v, -1)
	assert.Nil(t, err)
	assert.True(t, ok)

	get, err := client.Get(context.Background(), k)
	assert.Nil(t, err)
	assert.EqualValues(t, v, get)
	t.Log(string(get))
}

func TestClient_TTL(t *testing.T) {
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
	defer client.Close()

	k, v := []byte("key"), []byte("value")
	ok, err := client.Put(context.Background(), k, v, -1)
	assert.Nil(t, err)
	assert.True(t, ok)

	get, err := client.TTL(context.Background(), k)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, get)
}

func TestClient_Del(t *testing.T) {
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
	defer client.Close()

	k, v := []byte("key"), []byte("value")
	ok, err := client.Put(context.Background(), k, v, -1)
	assert.Nil(t, err)
	assert.True(t, ok)

	get, err := client.Get(context.Background(), k)
	assert.Nil(t, err)
	assert.EqualValues(t, v, get)
	t.Log(string(get))

	delOk, err := client.Del(context.Background(), k)
	assert.Nil(t, err)
	assert.True(t, delOk)

	get, err = client.Get(context.Background(), k)
	assert.ErrorIs(t, err, riverdb.ErrKeyNotFound)
}

func TestClient_Exp(t *testing.T) {
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
	defer client.Close()

	k, v := []byte("key"), []byte("value")
	ok, err := client.Put(context.Background(), k, v, -1)
	assert.Nil(t, err)
	assert.True(t, ok)

	get, err := client.Get(context.Background(), k)
	assert.Nil(t, err)
	assert.EqualValues(t, v, get)
	t.Log(string(get))

	expOk, err := client.Exp(context.Background(), k, time.Second)
	assert.Nil(t, err)
	assert.True(t, expOk)

	time.Sleep(time.Second * 2)

	get, err = client.Get(context.Background(), k)
	assert.ErrorIs(t, err, riverdb.ErrKeyNotFound)
}

func TestClient_Stat(t *testing.T) {
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
	defer client.Close()

	k, v := []byte("key"), []byte("value")
	ok, err := client.Put(context.Background(), k, v, -1)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = client.Put(context.Background(), k, v, -1)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = client.Put(context.Background(), k, v, -1)
	assert.Nil(t, err)
	assert.True(t, ok)

	stat, err := client.Stat(context.Background())
	assert.Nil(t, err)
	t.Log(stat)
}

func TestClient_PutInBatch(t *testing.T) {
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
	defer client.Close()

	var rs []*riverpb.Record
	for i := 0; i < 10; i++ {
		rs = append(rs, &riverpb.Record{
			Key:   []byte(fmt.Sprintf("a-%d", i)),
			Value: []byte("hello world"),
			Ttl:   0,
		})
	}

	// put all
	effected, err := client.PutInBatch(context.Background(), rs, int64(len(rs)))
	assert.Nil(t, err)
	assert.EqualValues(t, int64(len(rs)), effected)

	for _, r := range rs {
		value, err := client.Get(context.Background(), r.Key)
		assert.Nil(t, err)
		assert.EqualValues(t, r.Value, value)
	}

	var ks [][]byte
	for _, r := range rs {
		ks = append(ks, r.Key)
	}

	// delete all
	effected, err = client.DelInBatch(context.Background(), ks, int64(len(ks)))
	assert.Nil(t, err)
	assert.EqualValues(t, int64(len(ks)), effected)

	for _, r := range rs {
		_, err := client.Get(context.Background(), r.Key)
		assert.ErrorIs(t, err, riverdb.ErrKeyNotFound)
	}
}
