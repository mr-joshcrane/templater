package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mr-joshcrane/templater"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "takes no arguments, run in the PROJECT folder and make sure CSV files are present")
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	var files []string
	dir, _ := os.ReadDir(workingDir)
	for _, file := range dir {
		if strings.HasSuffix(file.Name(), ".csv") {
			p := path.Join(workingDir, file.Name())
			files = append(files, p)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = templater.GenerateTemplate(files)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
