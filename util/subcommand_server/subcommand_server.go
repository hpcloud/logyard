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
		return
	}

	var request map[string]string
	if err := json.Unmarshal(body, &request); err != nil {
		l.Errorf("Failed to decode JSON body in POST request (%s). Original body was: %s", err, string(body))
		return
	}

	// TODO: invoke the subcommand based on json params.
	fmt.Fprintf(w, "%+v", request)
}
