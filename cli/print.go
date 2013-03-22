package cli

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"github.com/wsxiaoys/terminal/color"
)

// Print a message from logyard streams
func PrintMessage(msg []byte) {
	var record map[string]interface{}

	if err := json.Unmarshal(msg, &record); err != nil {
		log.Fatal(err)
	}

	// REFACTOR: use Go interfaces to moves these to the respective
	// packages (apptail, cloudevents, systail).
	if _, ok := record["Name"]; ok {
		// systail
		name := record["Name"].(string)
		nodeid := record["NodeID"].(string)
		text := record["Text"].(string)

		color.Printf("@y%s@|@@@c%s@|: %s\n", name, nodeid, text)
	} else if _, ok = record["Type"]; ok {
		// cloud event
		kind := record["Type"].(string)
		nodeid := record["NodeID"].(string)
		severity := record["Severity"].(string)
		process := record["Process"].(string)
		desc := record["Desc"].(string)

		color.Printf("@g%s[%s]@|::@y%s@!@@@c%s@|: %s\n",
			kind, severity, process, nodeid, desc)
	} else if _, ok = record["Source"]; ok {
		// app logs
		appname := record["AppName"].(string)
		nodeid := record["NodeID"].(string)
		source := record["Source"].(string)
		text := record["Text"].(string)

		color.Printf("@b%s[%s]@|@@@c%s@|: %s\n",
			appname, source, nodeid, text)

	}
}
