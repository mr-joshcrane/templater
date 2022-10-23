package templater

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"cuelang.org/go/cue"
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
		tableName := filepath.Base(path)
		tableName = strings.ToUpper(tableName)
		tableName = strings.ReplaceAll(tableName, ".CSV", "")

		expr, err := json.Extract("", contents)
		if err != nil {
			fmt.Println(err)
			return errors.New("unable to convert json to cue")
		}

		v := c.BuildExpr(expr)

		table := Table{
			Name:        tableName,
			Fields:      make(map[string]Field),
			Project:     projectName,
			SQLTemplate: SQLTemplate{},
		}

		tables = append(tables, &table)

		item, err := v.List()
		if err != nil {
			panic(err)
		}
		for item.Next() {
			v1 := item.Value().LookupPath(cue.ParsePath("V"))
			if v1.Exists() {
				byt, err := v1.Bytes()
				if err != nil {
					panic(err)
				}
				e, err := json.Extract("", byt)
				if err != nil {
					panic(err)
				}
				v2 := c.BuildExpr(e)
				v2.Walk(stopCondition, func(c cue.Value) { unpack(&table, c, prefix) })
			}

			item.Value().Walk(func(c cue.Value) bool { return true }, func(c cue.Value) { unpack(&table, c, func(s string) string { return strings.ReplaceAll(s, `"`, ``) }) })
			if len(table.Fields) == 0 {
				return errors.New("empty JSON")
			}

		}
		delete(table.Fields, "V")

		err = GenerateSQLModel(table)
		if err != nil {
			return err
		}

		model := GenerateModel(tables)

		err = WriteModelProperties("transform_schema.yml", c, model)
		if err != nil {
			return err
		}
		err = WriteModelProperties("public_schema.yml", c, *model.AddDescriptions())
		if err != nil {
			return err
		}
		err = WriteSourceProperties("source_properties.yml", c, tables, projectName)
		if err != nil {
			return err
		}
	}
	return nil
}

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == '_' ||
			b == ' ' ||
			b == '.' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func formatKey(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, `(`, " ")
	s = strip(s)
	s = strings.TrimLeft(s, ` `)
	s = strings.ReplaceAll(s, `.`, `__`)
	s = strings.ReplaceAll(s, ` `, `_`)
	return s
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

func stopCondition(c cue.Value) bool {
	exp, err := regexp.Compile(`^[[0-9]*].`)
	if err != nil {
		panic(err)
	}
	p := c.Path().String()
	p = exp.ReplaceAllString(p, "")
	if strings.Contains(p, `[`) {
		return false
	}
	return true
}

type NameOption func(string) string

func prefix(s string) string {
	return "V:" + s
}

func unpack(t *Table, c cue.Value, opts ...NameOption) {
	exp, err := regexp.Compile(`^[[0-9]*].`)
	if err != nil {
		panic(err)
	}
	path := c.Path().String()
	path = exp.ReplaceAllString(path, "")
	node := path
	node = formatKey(node)

	for _, opt := range opts {
		path = opt(path)
	}
	path = strings.ReplaceAll(path, `"`, "")
	path = strings.ReplaceAll(path, `:`, `":"`)
	path = strings.ReplaceAll(path, `.`, `"."`)

	snowflakeType := SnowflakeTypes[c.IncompleteKind().String()]
	if strings.Contains(path, `[`) {
		return
	}
	if snowflakeType == "OBJECT" {
		return
	}

	field := Field{
		Node:        node,
		Path:        path,
		InferedType: snowflakeType,
	}

	if _, ok := t.Fields[path]; !ok {
		t.Fields[path] = field
		return
	}
	existingField := t.Fields[path]
	if existingField.InferedType == "VARCHAR" || existingField.InferedType == "" {
		existingField.InferedType = snowflakeType
	}
}
