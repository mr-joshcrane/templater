package templater

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
)

type Field struct {
	Node         string
	Path         string
	InferredType string
}

type Table struct {
	Name        string
	Project     string
	Fields      map[string]Field
	rawContents io.Reader
}

func GenerateProject(fsys fs.FS, projectName string, unpackPaths ...string) error {
	c := cuecontext.New()
	tables, err := GenerateTables(fsys, projectName, unpackPaths...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		err := GenerateTableFields(table, c, unpackPaths...)
		if err != nil {
			return err
		}
		writeTableModel(table)
	}

	models := GenerateProjectModel(tables)
	sources := GenerateProjectSources(tables, projectName)

	return WriteProperties(c, models, sources)
}

func createProjectDirectories() error {
	_, err := os.Stat("output")
	if err != nil {
		err := os.Mkdir("output", os.ModePerm)
		if err != nil {
			return err
		}
	}
	_, err = os.Stat("output/transform")
	if err != nil {
		err := os.Mkdir("output/transform", os.ModePerm)
		if err != nil {
			return err
		}
	}
	_, err = os.Stat("output/public")
	if err != nil {
		err := os.Mkdir("output/public", os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func Main() int {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	fsys := os.DirFS(workingDir)
	projectName := filepath.Base(workingDir)
	unpackFields := os.Args[1:]

	err = createProjectDirectories()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	err = GenerateProject(fsys, projectName, unpackFields...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
