package subcommand_server

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/logyard/util/subcommand"
)

// UserRequest is a request to run a specific sub-command with args.
type UserRequest struct {
	SubCommandName string   `json:"subcommand"`
	Arguments      []string `json:"arguments"`
}

func NewUserRequest(postBody []byte) (*UserRequest, error) {
	params := new(UserRequest)
	if err := json.Unmarshal(postBody, params); err != nil {
		return nil, fmt.Errorf("Invalid or incorrect JSON in POST (%s)", err)
	}

	// User must not pass -json
	for _, arg := range params.Arguments {
		if arg == "-json" {
			return nil, fmt.Errorf("Cannot pass -json")
		}
	}

	// Prepend -json to Arguments
	params.Arguments = append([]string{"-json"}, params.Arguments...)

	return params, nil
}

// Run runs the underlying subcommand, matching with the given
// subcommands list.
func (r *UserRequest) Run(cmds []subcommand.SubCommand) (string, error, error) {
	for _, sc := range cmds {
		if sc.Name() == r.SubCommandName {
			fs := subcommand.NewSubCommandFlagSet(sc)
			output, err := fs.ParseAndRun(r.Arguments)
			return output, err, nil
		}
	}

	err := fmt.Errorf("Invalid subcommand '%s'", r.SubCommandName)
	return "", nil, err
}
