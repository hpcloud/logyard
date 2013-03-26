package stream

import (
	"github.com/ActiveState/log"
	"logyard/util/pubsub"
	"strings"
)

func Stream(ch chan []byte, options MessagePrinterOptions) {
	// REFACTOR: do we need MessagePrinter at all? all it does to
	// provide abstraction over color formatting; most other things
	// (formatting, skipping) happens in stream_handler.go.
	printer := NewMessagePrinter(options)
	printer.AddFormat("systail",
		"@m{{.Name}}@|@@@c{{.NodeID}}@|: {{.Text}}")
	printer.AddFormat("event",
		"@g{{.Type}}@|[@m{{.Process}}@|]@@@c{{.NodeID}}@|: {{.Desc}}")
	printer.AddFormat("apptail",
		"@b{{.AppName}}[{{.Source}}]@|@@@c{{.NodeID}}@|: {{.Text}}")

	printer.SetPrePrintHook(streamHandler)

	// Print incoming records
	for line := range ch {
		parts := strings.SplitN(string(line), " ", 2)
		if len(parts) != 2 {
			log.Fatal("received invalid message: %s", string(line))
		}
		msg := pubsub.Message{parts[0], parts[1]}
		if err := printer.Print(msg); err != nil {
			log.Fatalf("Error -- %s -- printing message %v", err, msg)
		}
	}
}
