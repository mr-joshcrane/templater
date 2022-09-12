package main

import (
	"fmt"
	"os"

	"github.com/mr-joshcrane/templater"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "require 3 arguments, filepath, project, table")
		os.Exit(1)
	}
	filePath := os.Args[1]
	project := os.Args[2]
	table := os.Args[3]
	template, err := templater.GenerateTemplate(filePath, project, table)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println(template)
}