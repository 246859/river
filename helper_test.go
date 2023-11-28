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

func TestMask(t *testing.T) {
	const (
		A = 1 << iota
		B
		C
		D
		E
		F
	)
	var m Mask
	m.Store(A, B, E, F)

	res1 := m.CheckAny(A, E, D)
	assert.True(t, res1)

	res2 := m.CheckSome(A, B)
	assert.True(t, res2)

	m.Remove(B)

	res3 := m.CheckAny(B)
	assert.False(t, res3)
}
