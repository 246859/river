package riverdb

import (
	"crypto/sha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemHash(t *testing.T) {
	s := []byte("hello world!")
	hash1 := memHash(s)
	hash2 := memHash(s)
	assert.Equal(t, hash1, hash2)
	sha1.Sum(s)
}

func TestBitFlag(t *testing.T) {
	const (
		A = 1 << iota
		B
		C
		D
		E
		F
	)
	var m BitFlag
	m.Store(A, B, E, F)

	res1 := m.Check(A, E, D)
	assert.True(t, res1)

	m.Revoke(B)

	res3 := m.Check(B)
	assert.False(t, res3)
}
