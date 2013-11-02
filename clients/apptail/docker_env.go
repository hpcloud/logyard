package apptail

import (
	"encoding/json"
	"path/filepath"
)

func GetDockerImageEnv(rootPath string) (map[string]string, error) {
	data, err := ReadFileLimit(filepath.Join(rootPath, "/etc/stackato/image.env.json"), 50*1000)
	if err != nil {
		return nil, err
	}

	env := map[string]string{}

	err = json.Unmarshal(data, &env)
	return env, err
}
