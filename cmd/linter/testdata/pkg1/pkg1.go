package pkg1

import (
	"log"
	"os"
)

func main() {
	log.Fatal("error") // want "call to log.Fatal outside main.main"

	os.Exit(0) // want "call to os.Exit outside main.main"

	panic("error") // want "use of built-in panic"
}
