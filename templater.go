package templater

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"github.com/go-gota/gota/dataframe"
)

type Field struct {
	Node         string
	Path         string
	InferredType string
}

type Table struct {
	Name    string
	Project string
	Fields  map[string]Field
}

func GenerateTemplateFiles(filePaths []string) error {
	c := cuecontext.New()
	projectName := filepath.Dir(filePaths[0])
	projectName = filepath.Base(projectName)

	tables := []*Table{}

	for _, path := range filePaths {
		buf := bytes.NewBuffer([]byte{})
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		reader := bytes.NewReader(contents)
		df := dataframe.ReadCSV(reader, dataframe.WithLazyQuotes(true))
		err = df.WriteJSON(buf)
		if err != nil {
			return err
		}
		cueValue := c.CompileBytes(buf.Bytes())

		tableName := cleanTableName(path)

		table, err := MakeTable(cueValue, tableName, projectName, "V")
		if err != nil {
			return err
		}

		tables = append(tables, &table)

		filename := fmt.Sprintf("output/%s.sql", table.Name)
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		err = WriteSQLModel(table, file)
		if err != nil {
			return err
		}
	}

	models := GenerateModel(tables)
	sources := generateSources(tables, projectName)

	return WriteProperties(c, models, sources)
}

func Main() int {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "takes no arguments, run in the PROJECT folder and make sure CSV files are present")
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	var files []string
	dir, err := os.ReadDir(workingDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	_, err = os.Stat("output")
	if err != nil {
		err := os.Mkdir("output", os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
	}

	for _, file := range dir {
		if strings.HasSuffix(file.Name(), ".csv") {
			p := path.Join(workingDir, file.Name())
			files = append(files, p)
		}
	}
	err = GenerateTemplateFiles(files)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
