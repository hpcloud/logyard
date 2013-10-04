package stream

import (
	"github.com/ActiveState/logyard/util/golor"
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
		// auth.log prefix
		regexp.MustCompile(
			`\w+ \d+ \d+\:\d+\:\d+ (.+)`),
	}
}

func handleSystail(record map[string]interface{}, options MessagePrinterOptions) bool {
	text := record["Text"].(string)
	process := record["Name"].(string)
	node := record["NodeID"].(string)
	severity := ""

	if len(options.NodeID) > 0 && node != options.NodeID {
		return false
	}

	if !options.LogyardVerbose && process == "github.com/ActiveState/logyard" && strings.Contains(text, "INFO") {
		return false
	}

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

	if !options.NoColor {
		record["NodeID"] = golor.Colorize(node, golor.GRAY, -1)
		switch severity {
		case "ERROR":
			record["Text"] = golor.Colorize(text, golor.RED, -1)
		case "WARN":
			// yellow
			record["Text"] = golor.Colorize(text, golor.YELLOW, -1)
		default:
			record["Text"] = text
		}

		// Assign an unique color to the process name
		record["Name"] = golor.Colorize(process, golor.AssignColor(process), -1)
	}
	return true
}

func handleEvent(record map[string]interface{}, options MessagePrinterOptions) bool {
	desc := record["Desc"].(string)
	severity := record["Severity"].(string)
	node := record["NodeID"].(string)
	typ := record["Type"].(string)
	process := record["Process"].(string)

	if len(options.NodeID) > 0 && node != options.NodeID {
		return false
	}

	if !options.NoColor {
		record["NodeID"] = golor.Colorize(node, golor.GRAY, -1)
		record["Type"] = golor.Colorize(typ, golor.MAGENTA, -1)
		record["Process"] = golor.Colorize(process, golor.BLUE, -1)
		switch severity {
		case "ERROR":
			record["Desc"] = golor.Colorize(desc, -1, golor.RED)
		case "WARNING":
			record["Desc"] = golor.Colorize(desc, 0, golor.YELLOW)
		default:
		}
	}
	return true
}

func handleApptail(record map[string]interface{}, options MessagePrinterOptions) bool {
	appname := record["AppName"].(string)
	node := record["NodeID"].(string)

	if len(options.NodeID) > 0 && node != options.NodeID {
		return false
	}

	if !options.NoColor {
		record["NodeID"] = golor.Colorize(node, golor.GRAY, -1)
		record["AppName"] = golor.Colorize(appname, golor.BLUE, -1)
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
	} else if keypart1 == "apptail" {
		return handleApptail(record, options)
	}

	return true
}
