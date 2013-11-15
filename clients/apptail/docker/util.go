package docker

import (
	"io"
	"io/ioutil"
	"os"
)

// readFileLimit is like ioutil.ReadFile, but with a read limit set.
func readFileLimit(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := io.LimitReader(file, limit)
	return ioutil.ReadAll(r)
}
