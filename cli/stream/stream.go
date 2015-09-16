package stream

import (
	"fmt"
	"github.com/hpcloud/log"
	"github.com/hpcloud/zmqpubsub"
	"strings"
)

func Stream(ch chan string, options MessagePrinterOptions) {
	// XXX: do we need MessagePrinter at all? all it does is
	// provide abstraction over color formatting; most other things
	// (formatting, skipping) happen in handler.go.
	printer := NewMessagePrinter(options)

	printer.AddFormat("systail",
		"{{.name}}@{{.node_id}}: {{.text}}")
	printer.AddFormat("event",
		"{{.type}}[{{.process}}]@{{.node_id}}: {{.desc}}")
	printer.AddFormat("apptail",
		"{{.app_name}}[{{.source}}]@{{.node_id}}: {{.text}}")

	printer.SetPrePrintHook(streamHandler)

	// Print incoming records
	for line := range ch {
		parts := strings.SplitN(string(line), " ", 2)
		if len(parts) != 2 {
			printer.PrintInternalError(fmt.Sprintf(
				"received invalid message: %v", string(line)))
			continue
		}
		msg := zmqpubsub.Message{parts[0], parts[1]}
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
