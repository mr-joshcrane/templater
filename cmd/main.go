package main

import (
	"fmt"
	"os"
	"strings"

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

	var contents []byte
	var err error

	contents, err = os.ReadFile(filePath)
	if err == nil && strings.HasSuffix(filePath, ".csv") {
		contents, err = templater.CsvToJson(contents)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	template, err := templater.GenerateTemplate(contents, project, table)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println(template)
}
