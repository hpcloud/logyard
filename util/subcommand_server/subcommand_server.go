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

type SubCommandServer struct {
	Commands []subcommand.SubCommand
}

func (srv SubCommandServer) Start(port int) {
	http.HandleFunc("/", srv.Handler)
	log.Fatal(
		http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil))
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

	var params Params
	if err := json.Unmarshal(body, &params); err != nil {
		l.Errorf("Failed to decode JSON body in POST request (%s). Original body was: %s", err, string(body))
		http.Error(w, err.Error(), 500)
		return
	}

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
