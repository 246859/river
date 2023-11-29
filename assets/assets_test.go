package assets

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBanner(t *testing.T) {
	err := OutFsFile(Fs, "banner.txt", os.Stdout)
	assert.Nil(t, err)
}
