package main

import (
	"log"
	"os"
)

func main() {
	// OK: разрешено в main.main
	log.Fatal("exit")
	os.Exit(0)
}

func other() {
	log.Fatal("error") // want "call to log.Fatal outside main.main"

	os.Exit(0) // want "call to os.Exit outside main.main"
}
