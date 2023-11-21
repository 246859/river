package riverdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxnHeap(t *testing.T) {
	data := []*Txn{
		{startedTs: 1},
		{startedTs: 4},
		{startedTs: -1},
		{startedTs: 100},
		{startedTs: 23},
		{startedTs: 7},
	}
	th := makeTxnPriorityHeap()

	for _, datum := range data {
		th.push(datum)
	}
	assert.EqualValues(t, -1, th.pop().startedTs)
	th.remove(data[0])
	assert.EqualValues(t, 4, th.pop().startedTs)
}
