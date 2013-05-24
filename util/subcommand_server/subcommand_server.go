// subcommand_server exposes subcommands defined in the 'subcommand'
// package as a HTTP server.
package subcommand_server

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"io/ioutil"
	"logyard/util/subcommand"
	"net/http"
)

type Params struct {
	SubCommandName string   `json:"subcommand"`
	Arguments      []string `json:"arguments"`
}

func NewParams(postBody []byte) (*Params, error) {
	params := new(Params)
	if err := json.Unmarshal(postBody, params); err != nil {
		return nil, fmt.Errorf("Invalid JSON in POST (%s)", err)
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

type SubCommandServer struct {
	Commands []subcommand.SubCommand
}

func (srv SubCommandServer) Start(addr string) error {
	http.HandleFunc("/", srv.Handler)
	return http.ListenAndServe(addr, nil)
}

func (srv SubCommandServer) Handler(w http.ResponseWriter, r *http.Request) {
	l := log.New()
	l.SetPrefix(fmt.Sprintf("[HTTP:%p]", r))

	l.Info("%+v", r)

	// TODO: replace http.Error, specifically err.Error(), with JSON
	// wrapped response writer.

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Error(err)
		http.Error(w, err.Error(), 500)
		return
	}

	params, err := NewParams(body)
	if err != nil {
		l.Error(err)
		http.Error(w, err.Error(), 400)
		return
	}

	l.Infof("Invoking: %s %+v", params.SubCommandName, params.Arguments)

	for _, sc := range srv.Commands {
		if sc.Name() == params.SubCommandName {
			fs := subcommand.NewSubCommandFlagSet(sc)
			output, err := fs.ParseAndRun(params.Arguments)
			if err != nil {
				l.Error(err)
				http.Error(w, err.Error(), 500)
			} else {
				if _, err := w.Write([]byte(output)); err != nil {
					l.Error(err)
				}
			}
			return
		}
	}

	err = fmt.Errorf("Invalid subcommand '%s'", params.SubCommandName)
	l.Error(err)
	http.Error(w, err.Error(), 400)
}
