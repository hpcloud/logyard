// 256 color terminal printing in Go.
package xtermcolor

import (
	"fmt"
)

func Colorize(s string, fg int) string {
	if fg < 0 || fg > 255 {
		panic("invalid color index")
	}
	escape := fmt.Sprintf("\033[%sm", colorCode(fg))
	escape += fmt.Sprintf("%s\033[0m", s)
	return escape
}

func colorCode(index int) string {
	return fmt.Sprintf("38;05;%d", index)
}
