package stream

import (
	"fmt"
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
			printer.PrintInternalError(fmt.Sprintf(
				"received invalid message: %v", string(line)))
			continue
		}
		msg := pubsub.Message{parts[0], parts[1]}
		if !(strings.HasPrefix(msg.Key, "systail") ||
			strings.HasPrefix(msg.Key, "apptail") ||
			strings.HasPrefix(msg.Key, "event")) {
			printer.PrintInternalError(fmt.Sprintf(
				"unsupported stream key (%s) for message: %v",
				msg.Key, msg.Value))
			continue
		}
		if err := printer.Print(msg); err != nil {
			log.Fatalf("Error -- %s -- printing message %s:%s",
				err, msg.Key, msg.Value)
		}
	}
}
