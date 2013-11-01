package apptail

import (
	_ "fmt"
	"io"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
)

type StackatoYml struct {
	Name     string            `yaml:"name"`
	LogFiles map[string]string `yaml:"logfiles"`
}

func NewStackatoYml(path string) (*StackatoYml, error) {
	data, err := getStackatoYmlSecure(path)
	if err != nil {
		return nil, err
	}

	s := new(StackatoYml)

	goyaml.Unmarshal(data, s)

	// fmt.Printf("Unmarshal(%v) => %+v\n", string(data), s)

	return s, nil
}

func getStackatoYmlSecure(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := io.LimitReader(file, 50*100)
	return ioutil.ReadAll(r)
}
