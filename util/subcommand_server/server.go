// subcommand_server exposes subcommands defined in the 'subcommand'
// package as a HTTP server.
package subcommand_server

import (
	"fmt"
	"github.com/ActiveState/log"
	"io/ioutil"
	"logyard/util/subcommand"
	"net/http"
)

type Server struct {
	Commands []subcommand.SubCommand
}

func (srv Server) Start(addr string) error {
	http.HandleFunc("/", srv.Handler)
	return http.ListenAndServe(addr, nil)
}

func (srv Server) Handler(w http.ResponseWriter, r *http.Request) {
	l := log.New()
	l.SetPrefix(fmt.Sprintf("[HTTP:%p]", r))

	l.Info("%+v", r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Error(err)
		http.Error(w, err.Error(), 500)
		return
	}

	params, err := NewUserRequest(body)
	if err != nil {
		l.Error(err)
		http.Error(w, err.Error(), 400)
		return
	}

	l.Infof("Invoking: %s %+v", params.SubCommandName, params.Arguments)

	output, cmdErr, err := params.Run(srv.Commands)
	if err != nil {
		l.Error(err)
		http.Error(w, err.Error(), 400)
	} else {
		if cmdErr != nil {
			// XXX: perhaps we should wrap the error as a JSON object
			// just as we do for 200-code responses.
			l.Error(err)
			http.Error(w, err.Error(), 500)
		} else {
			// Command ran successfully. Send the output back.
			if _, err = w.Write([]byte(output)); err != nil {
				l.Error(err)
			}
		}
	}
}
