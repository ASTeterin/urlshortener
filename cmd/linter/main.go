package main

import (
	"fmt"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	fmt.Println("!!!!!!!!!!")
	singlechecker.Main(Analyzer)
}
