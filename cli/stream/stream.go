package stream

import (
	"github.com/ActiveState/log"
	"logyard/util/pubsub"
	"strings"
)

func Stream(ch chan []byte, options MessagePrinterOptions) {
	// XXX: do we need MessagePrinter at all? all it does is
	// provide abstraction over color formatting; most other things
	// (formatting, skipping) happen in handler.go.
	printer := NewMessagePrinter(options)

	printer.AddFormat("systail",
		"{{.Name}}@{{.NodeID}}: {{.Text}}")
	printer.AddFormat("event",
		"{{.Type}}[{{.Process}}]@{{.NodeID}}: {{.Desc}}")
	printer.AddFormat("apptail",
		"{{.AppName}}[{{.Source}}]@{{.NodeID}}: {{.Text}}")

	printer.SetPrePrintHook(streamHandler)

	// Print incoming records
	for line := range ch {
		parts := strings.SplitN(string(line), " ", 2)
		if len(parts) != 2 {
			log.Fatal("received invalid message: %s", string(line))
		}
		msg := pubsub.Message{parts[0], parts[1]}
		if err := printer.Print(msg); err != nil {
			log.Fatalf("Error -- %s -- printing message %s:%s",
				err, msg.Key, msg.Value)
		}
	}
}
