package docker

import (
	"io"
	"io/ioutil"
	"os"
)

// ReadFileLimit is like ioutil.ReadFile, but with a read limit set.
func ReadFileLimit(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := io.LimitReader(file, limit)
	return ioutil.ReadAll(r)
}
