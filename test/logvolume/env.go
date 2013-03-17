package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func handler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, strings.Join(os.Environ(), "\n"))
}

func main() {
	go func() {
		for _ = range time.Tick(100 * time.Millisecond) {
			log.Println("Tick!")
			fmt.Println("Tick on stdout!")
		}
	}()
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
