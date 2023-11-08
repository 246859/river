package assets

import (
	"github.com/246859/river/pkg/file"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBanner(t *testing.T) {
	err := file.OutFsFile(Fs, "banner.txt", os.Stdout)
	assert.Nil(t, err)
}
