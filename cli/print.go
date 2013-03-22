package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"logyard/util/pubsub"
	"strings"
	"text/template"
)

// MessagePrinter handles print representation of messages streamed by
// logyard.
type MessagePrinter struct {
	templates map[string]*template.Template
}

func NewMessagePrinter() MessagePrinter {
	return MessagePrinter{make(map[string]*template.Template)}
}

// Add print format for messages identified by this key prefix. The
// prefix of the key must not contain any period. For example, if
// messages are identified by "systail.dea.NODE", then keypart1 should
// just be "systail".
func (p MessagePrinter) AddFormat(keypart1 string, format string) {
	if _, ok := p.templates[keypart1]; ok {
		panic("already added")
	}
	p.templates[keypart1] = template.Must(
		template.New("print-" + keypart1).Parse(format))
}

// Print a message from logyard streams
func (p MessagePrinter) Print(msg pubsub.Message) error {
	var record map[string]interface{}

	if err := json.Unmarshal([]byte(msg.Value), &record); err != nil {
		return err
	}

	key := msg.Key
	if strings.Contains(key, ".") {
		key = strings.SplitN(key, ".", 2)[0]
	}

	if tmpl, ok := p.templates[key]; ok {
		var buf bytes.Buffer

		escapeSpecialColorChars(record)

		if err := tmpl.Execute(&buf, record); err != nil {
			return err
		}
		color.Println(string(buf.Bytes()))
		return nil
	}
	return fmt.Errorf("no format added for key: %s", key)
}

// escapeSpecialColorChars escapes special color chars from the string
// values in the map.
func escapeSpecialColorChars(m map[string]interface{}) {
	for key, value := range m {
		if s, ok := value.(string); ok {
			m[key] = strings.Replace(s, "@", "@@", -1)
		}
	}
}
