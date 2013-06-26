package stream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"logyard/util/golor"
	"logyard/util/pubsub"
	"strings"
	"text/template"
	"time"
)

// REFACTOR: remove coupling between color printer and options.
type MessagePrinterOptions struct {
	Raw            bool
	LogyardVerbose bool
	ShowTime       bool
	NoColor        bool
	NodeID         string
	JSON           bool
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
	p.templates[keypart1] = template.Must(
		template.New("print-" + keypart1).Parse(format))
}

func (p *MessagePrinter) SetPrePrintHook(fn FilterFn) {
	p.filterFn = fn
}

// Print a message from logyard streams
func (p MessagePrinter) Print(msg pubsub.Message) error {
	if p.options.JSON {
		key, value := msg.Key, msg.Value
		if !p.options.NoColor {
			key = golor.Colorize(msg.Key, golor.RGB(0, 4, 4), -1)
		}
		fmt.Printf("%s %s\n", key, value)
		return nil
	}

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

		if p.filterFn(key, record, p.options) {
			if err := tmpl.Execute(&buf, record); err != nil {
				return err
			}
			s := string(buf.Bytes())
			if p.options.ShowTime {
				s = fmt.Sprintf("%s %s", time.Now(), s)
			}
			fmt.Println(s)
		}
		return nil
	}
	return fmt.Errorf("no format added for key: %s", key)
}
