package main

import (
	"fmt"
	"logyard/util/xtermcolor"
)

func main() {
	for i := 255; i >= 0; i-- {
		fmt.Println(xtermcolor.Colorize("Foreground", i, -1))
		// fmt.Println(xtermcolor.ColorizeBg("Background", i, 100))
	}
}
