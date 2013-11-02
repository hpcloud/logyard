package apptail

import (
	_ "fmt"
	"launchpad.net/goyaml"
)

type StackatoYml struct {
	Name     string            `yaml:"name"`
	LogFiles map[string]string `yaml:"logfiles"`
}

func NewStackatoYml(path string) (*StackatoYml, error) {
	data, err := ReadFileLimit(path, 50*100)
	if err != nil {
		return nil, err
	}

	s := new(StackatoYml)

	goyaml.Unmarshal(data, s)

	// fmt.Printf("Unmarshal(%v) => %+v\n", string(data), s)

	return s, nil
}
