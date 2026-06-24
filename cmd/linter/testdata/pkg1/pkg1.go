package pkg1

import (
	"log"
	"os"
)

func main() {
	// want "call to log.Fatal outside main.main"
	log.Fatal("error")

	// want "call to os.Exit outside main.main"
	os.Exit(0)

	// want "use of built-in panic"
	panic("error")
}
