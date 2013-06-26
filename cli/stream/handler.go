package stream

import (
	"crypto/sha1"
	"logyard/util/xtermcolor"
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

	if !options.LogyardVerbose && process == "logyard" && strings.Contains(text, "INFO") {
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
		record["NodeID"] = xtermcolor.Colorize(node, xtermcolor.RGB(0, 3, 3), -1)
		switch severity {
		case "ERROR":
			record["Text"] = xtermcolor.Colorize(text, xtermcolor.RGB(5, 0, 0), -1)
		case "WARN":
			// yellow
			record["Text"] = xtermcolor.Colorize(text, xtermcolor.RGB(5, 5, 0), -1)
		default:
			record["Text"] = text
		}

		// Assign an unique color to the process name
		record["Name"] = colorizeString(process)
	}
	return true
}

func handleEvent(record map[string]interface{}, options MessagePrinterOptions) bool {
	desc := record["Desc"].(string)
	severity := record["Severity"].(string)
	node := record["NodeID"].(string)

	if len(options.NodeID) > 0 && node != options.NodeID {
		return false
	}

	if !options.NoColor {
		record["NodeID"] = xtermcolor.Colorize(node, xtermcolor.RGB(0, 3, 3), -1)
		switch severity {
		case "ERROR":
			record["Desc"] = xtermcolor.Colorize(desc, -1, xtermcolor.RGB(5, 0, 0))
		case "WARNING":
			record["Desc"] = xtermcolor.Colorize(desc, 0, xtermcolor.RGB(5, 5, 0))
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
		record["NodeID"] = xtermcolor.Colorize(node, xtermcolor.RGB(0, 3, 3), -1)
		record["AppName"] = xtermcolor.Colorize(appname, xtermcolor.RGB(0, 0, 5), -1)
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

// Return the given string colorized to an unique value.
func colorizeString(s string) string {
	maxColor := 211 // prevent whitish colors; use prime number for mod.
	minColor := 17  // avoid 16 colors, including the black.
	fg := minColor + stringId(s, maxColor-minColor+1)
	return xtermcolor.Colorize(s, fg, -1)
}

func stringId(s string, mod int) int {
	h := sha1.New()
	h.Write([]byte(s))
	sum := 0
	for _, n := range h.Sum(nil) {
		sum += int(n)
	}
	return sum % mod
}
