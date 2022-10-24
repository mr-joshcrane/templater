package templater

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/json"
)

var SnowflakeTypes = map[string]string{
	"string": "STRING",
	"int":    "INTEGER",
	"float":  "FLOAT",
	"struct": "OBJECT",
	"list":   "ARRAY",
	"null":   "VARCHAR",
	"bool":   "BOOLEAN",
}

type Field struct {
	Node        string
	Path        string
	InferedType string
}

type Table struct {
	Name        string
	Project     string
	Fields      map[string]Field
	SQLTemplate SQLTemplate
}

func GenerateTemplate(filePaths []string) error {
	c := cuecontext.New()
	projectName := filepath.Dir(filePaths[0])
	projectName = filepath.Base(projectName)

	tables := []*Table{}

	for _, path := range filePaths {
		contents, err := os.ReadFile(path)
		if err == nil {
			contents, err = CsvToJson(contents)
			if err != nil {
				return err
			}
		}
		tableName := cleanTableName(path)

		expr, err := json.Extract("", contents)
		if err != nil {
			fmt.Println(err)
			return errors.New("unable to convert json to cue")
		}

		v := c.BuildExpr(expr)

		table, err := MakeTable(v, tableName, projectName)
		if err != nil {
			return err
		}

		tables = append(tables, &table)

		err = GenerateSQLModel(table)
		if err != nil {
			return err
		}
	}
	
	models := GenerateModel(tables)

	err := WriteModelProperties("transform_schema.yml", c, models)
	if err != nil {
		return err
	}
	err = WriteModelProperties("public_schema.yml", c, *models.AddDescriptions())
	if err != nil {
		return err
	}
	err = WriteSourceProperties("source_schema.yml", c, tables, projectName)
	if err != nil {
		return err
	}
	return nil
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
		err := os.Mkdir("output", 0777)
		if err != nil {
			return 1
		}
	}

	for _, file := range dir {
		if strings.HasSuffix(file.Name(), ".csv") {
			p := path.Join(workingDir, file.Name())
			files = append(files, p)
		}
	}
	err = GenerateTemplate(files)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
