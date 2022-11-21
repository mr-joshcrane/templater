package templater

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
)

// A Field represents a column and information about how it should be transformed.
// Node: Represents the post-transformation target identifier in Snowflake.
// Path: Represents the pre-transformation path to the data in the source table.
// InferType: Represents the current best guess at Snowflake type inferred from exemplars.
type Field struct {
	Node         string
	Path         string
	InferredType string
}

// A Table represents a source table.
// It is the intermediate representation of the untyped semi-structured data.
type Table struct {
	Name        string
	Project     string
	Fields      map[string]Field
	rawContents io.Reader
}

// generateProject given a [fs.FS] of CSV's and an optional list of fields to unpack, will generate the project.
func generateProject(fsys fs.FS, projectName string, unpackPaths ...string) error {
	c := cuecontext.New()
	tables, err := generateTables(fsys, projectName, unpackPaths...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		err := generateTableFields(table, c, unpackPaths...)
		if err != nil {
			return err
		}
	}

	models := GenerateProjectModel(tables)
	sources := generateProjectSources(tables, projectName)

	return writeProject(c, models, sources, tables)
}

// createProjectDirectories will create the neccessary project directories.
// This is a noop if the directories already exist.
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

// Main is the entrypoint for the templater.
// Working in the context of the current working directory as a [fs.FS]
// and taking [os.Args] as a list of fields to unpack
// it will generate a the following artifacts:
// - A [transform] directory containing the DBT SQL transformations of the tables.
// - A [public] directory containing the DBT SQL clone transforms.
// - Schemas for source, transform, and public models.
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

	err = generateProject(fsys, projectName, unpackFields...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
