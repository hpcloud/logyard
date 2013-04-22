package stream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"logyard/util/pubsub"
	"strings"
	"text/template"
	"time"
)

type MessagePrinterOptions struct {
	Raw      bool
	ShowTime bool
	NoColor  bool
}

// FilterFn is a function to filter incoming messages
type FilterFn func(
	keypart1 string,
	record map[string]interface{},
	options MessagePrinterOptions) bool

// MessagePrinter handles print representation of messages streamed by
// logyard.
type MessagePrinter struct {
	templates map[string]*template.Template
	options   MessagePrinterOptions
	filterFn  FilterFn
}

func NewMessagePrinter(options MessagePrinterOptions) MessagePrinter {
	return MessagePrinter{
		make(map[string]*template.Template), options, nil}
}

// Add print format for messages identified by this key prefix. The
// prefix of the key must not contain any period. For example, if
// messages are identified by "systail.dea.NODE", then keypart1 should
// just be "systail".
func (p MessagePrinter) AddFormat(keypart1 string, format string) {
	if _, ok := p.templates[keypart1]; ok {
		panic("already added")
	}
	if p.options.NoColor {
		format = stripColor(format)
		fmt.Printf("Added format: %s\n", format)
	}
	p.templates[keypart1] = template.Must(
		template.New("print-" + keypart1).Parse(format))
}

func (p *MessagePrinter) SetPrePrintHook(fn FilterFn) {
	p.filterFn = fn
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

		if p.filterFn(key, record, p.options) {
			if err := tmpl.Execute(&buf, record); err != nil {
				return err
			}
			s := string(buf.Bytes())
			if p.options.ShowTime {
				s = fmt.Sprintf("%s %s", time.Now(), s)
			}
			color.Println(s)
		}
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

func stripColor(s string) string {
	var buf bytes.Buffer
	mode := false
	for _, c := range s {
		switch {
		case mode && c == '@':
			buf.WriteString("@@") // handle @
			mode = false
		case mode && c != '@':
			mode = false
			continue // ignore the special char
		case c == '@' && !mode:
			mode = true // ignore unescaped @
		default:
			buf.WriteRune(c)
			mode = false
		}
	}
	return buf.String()
}
