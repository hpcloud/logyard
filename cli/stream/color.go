package stream

import (
	"crypto/sha1"
	"fmt"
	"logyard/util/xtermcolor"
)

// colorize applies the given color on the string.
func colorize(s string, code string) string {
	return fmt.Sprintf("@%s%s@|", code, s)
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

// Return the given string colorized to an unique value.
func colorizeString(s string) string {
	return xtermcolor.Colorize(s, 1+stringId(s, 255))
}
