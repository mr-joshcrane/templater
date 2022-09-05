package main

import (
	"fmt"
	"os"

	"github.com/mr-joshcrane/templater"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "require 3 arguments, filepath, project, table\n")
		os.Exit(1)
	}
	filePath := os.Args[1]
	project := os.Args[2]
	table := os.Args[3]
	template := templater.GenerateTemplate(filePath, project, table)
	fmt.Println(template)
}