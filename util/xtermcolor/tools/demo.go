package main

import (
	"fmt"
	"logyard/util/xtermcolor"
)

func main() {
	for i := 0; i < 256; i++ {
		fmt.Println(xtermcolor.Colorize("Foreground", i))
		// fmt.Println(xtermcolor.ColorizeBg("Background", i, 100))
	}
}
