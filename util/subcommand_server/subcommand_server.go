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
			// TODO: make subcommand.Parse extensible enough to use
			// here. i.e.,
			// * allow parsing of single sub-command and custom args
			// * not use stdout/stderr, but json.

			// TODO: have the subcommand return, instead of printing
			// to console.

			// TODO: add --json for all subcommands.

			if output, err := sc.Run(params.Arguments); err != nil {
				l.Error(err)
				http.Error(w, err.Error(), 500)
			} else {
				// TODO: return response from subcommand
				response := []byte(output)
				if _, err := w.Write(response); err != nil {
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
