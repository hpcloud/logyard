package docker

import (
	"encoding/json"
	"path/filepath"
)

func GetDockerAppEnv(rootPath string) (map[string]string, error) {
	data, err := readFileLimit(filepath.Join(rootPath, "/home/stackato/etc/droplet.env.json"), 50*1000)
	if err != nil {
		return nil, err
	}

	env := map[string]string{}

	err = json.Unmarshal(data, &env)
	return env, err
}
