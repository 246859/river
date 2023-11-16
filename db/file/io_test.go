package file

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func Test(t *testing.T) {
	ios := []func(string) (IO, error){
		func(filename string) (IO, error) {
			return OpenStdFile(filename)
		},
	}

	for _, io := range ios {
		randName := filepath.Join(os.TempDir(), strconv.Itoa(int(rand.Int31())))
		randIO, err := io(randName)
		assert.NotNil(t, randIO)
		assert.Nil(t, err)

		testIO_Write(t, randIO)
		testIO_Read(t, randIO)
		testIO_Sync(t, randIO)
		testIO_Close(t, randIO)

		removeErr := os.Remove(randName)
		assert.Nil(t, removeErr)
	}
}

func testIO_Write(t *testing.T, io IO) {
	n, err := io.Write([]byte("hello world!"))
	assert.True(t, n > 0)
	assert.Nil(t, err)

	n1, err1 := io.Write([]byte("hello world!"))
	assert.True(t, n1 > 0)
	assert.Nil(t, err1)
}

func testIO_Read(t *testing.T, io IO) {
	bytes := []byte("hello world!")

	buf1 := make([]byte, len(bytes))
	n1, err1 := io.ReadAt(buf1, int64(len(bytes)))
	assert.True(t, n1 > 0)
	assert.Nil(t, err1)
	assert.Equal(t, bytes, buf1)

	buf2 := make([]byte, 0)
	n2, err2 := io.ReadAt(buf2, 0)
	assert.True(t, n2 == 0)
	assert.Nil(t, err2)
	assert.Equal(t, []byte{}, buf2)
}

func testIO_Sync(t *testing.T, io IO) {
	err := io.Sync()
	assert.Nil(t, err)
}

func testIO_Close(t *testing.T, io IO) {
	err := io.Close()
	assert.Nil(t, err)
}
