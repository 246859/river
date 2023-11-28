package riverdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
