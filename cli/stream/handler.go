package stream

import (
	"fmt"
	"regexp"
	"strings"
)

var datetimePatterns []*regexp.Regexp

func init() {
	datetimePatterns = []*regexp.Regexp{
		// vcap log prefix
		regexp.MustCompile(
			`\[[^\]]+\] \S+ \- \S+=\S+ \S+=\S+ \S+=\S+ (.+)`),
		// Go projects log prefix
		regexp.MustCompile(
			`\d+\/\d+\/\d+ \d+\:\d+\:\d+ (.+)`),
		// supervisord log prefix
		regexp.MustCompile(
			`\d+-\d+-\d+ \d+:\d+:\d+,\d+ (.+)`),
	}
}

func handleSystail(record map[string]interface{}, options MessagePrinterOptions) bool {
	text := record["Text"].(string)
	process := record["Name"].(string)
	severity := ""

	if process == "logyard" && strings.Contains(text, "INFO") {
		return false
	}

	// TODO: highlight process names
	// TODO: hide ip addr in micro cloud.

	if strings.Contains(process, "nginx") {
		// FIXME: nginx logs reflect requests to not only vcap
		// processes (eg: cc and services), but also deployed
		// apps. the purpose of `kato tail` is to tail the log of
		// vcap and other core processes only, not the deployed
		// apps. perhaps we could redirect app requests to a
		// different log file?
		switch {
		case strings.Contains(text, "[error]"):
			severity = "ERROR"
		case strings.Contains(text, "No such file or directory"):
			fallthrough
		case strings.Contains(text, "404"):
			severity = "WARN"
		default:
		}
	} else {
		switch {
		case strings.Contains(text, "ERROR"):
			severity = "ERROR"
		case strings.Contains(text, "WARN"):
			severity = "WARN"
		default:
		}
	}

	// Strip non-essential data
	if !options.Raw {
		for _, re := range datetimePatterns {
			res := re.FindStringSubmatch(text)
			if len(res) > 1 {
				text = res[1]
				break
			}
		}
	}

	switch severity {
	case "ERROR":
		record["Text"] = colorize(text, "r")
	case "WARN":
		record["Text"] = colorize(text, "y")
	default:
		record["Text"] = text
	}
	return true
}

func handleEvent(record map[string]interface{}, options MessagePrinterOptions) bool {
	desc := record["Desc"].(string)
	severity := record["Severity"].(string)

	switch severity {
	case "ERROR":
		record["Desc"] = colorize(desc, "R")
	case "WARNING":
		record["Desc"] = colorize(desc, "Y")
	default:
	}
	return true
}

func streamHandler(
	keypart1 string,
	record map[string]interface{},
	options MessagePrinterOptions) bool {

	if keypart1 == "systail" {
		return handleSystail(record, options)
	} else if keypart1 == "event" {
		return handleEvent(record, options)
	}

	return true
}

// colorize applies the given color on the string.
func colorize(s string, code string) string {
	return fmt.Sprintf("@%s%s@|", code, s)
}
