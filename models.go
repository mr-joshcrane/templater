package templater

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
	"golang.org/x/exp/maps"
)

var (
	//go:embed template.gohtml
	fs embed.FS
)

type SQLTemplate struct {
	Tags    string
	Columns string
	Source  string
}

func GenerateTagsSQL(project, table string) string {
	return fmt.Sprintf("{{ config(tags=['%s', '%s']) }}", strings.ToUpper(project), strings.ToUpper(table))
}

func GenerateSourceSQL(project, table string) string {
	return fmt.Sprintf("  {{ source('%s', '%s') }}", strings.ToUpper(project), strings.ToUpper(table))
}

func GenerateColumnsSQL(f map[string]Field) string {
	fields := maps.Values(f)
	column_data := ""
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].Node[0] == '_' && fields[j].Node[0] != '_' {
			return false
		}
		return fields[i].Node < fields[j].Node
	})
	for _, field := range fields {

		column_data += fmt.Sprintf(`  ,"%s"::%s AS %s`, field.Path, field.InferedType, NormaliseKey(field.Node))
		column_data += "\n"
	}
	// strip the first comma out
	column_data = strings.Replace(column_data, ",", "", 1)
	// strip the last new line out
	column_data = column_data[0 : len(column_data)-1]

	return column_data
}

func GenerateSQLModel(table Table) error {
	var body bytes.Buffer

	table.SQLTemplate = SQLTemplate{
		Tags:    GenerateTagsSQL(table.Project, table.Name),
		Columns: GenerateColumnsSQL(table.Fields),
		Source:  GenerateSourceSQL(table.Project, table.Name),
	}
	tpl, err := template.New("template.gohtml").ParseFS(fs, "template.gohtml")
	if err != nil {
		return err
	}
	err = tpl.Execute(&body, table.SQLTemplate)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("output/%s.sql", table.Name)
	err = os.WriteFile(filename, body.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func WriteModelProperties(path string, c *cue.Context, model Models) error {
	transformEncoded, err := yaml.Encode(c.Encode(model))
	if err != nil {
		return err
	}
	path = fmt.Sprintf("output/%s", path)
	err = os.WriteFile(path, transformEncoded, 0644)
	if err != nil {
		return err
	}
	return nil
}

func WriteSourceProperties(path string, c *cue.Context, tables []*Table, projectName string) error {
	sourceModel := generateSources(tables, projectName)
	sourceEncoded, err := yaml.Encode(c.Encode(sourceModel))
	if err != nil {
		return err
	}
	path = fmt.Sprintf("output/%s", path)

	err = os.WriteFile(path, sourceEncoded, 0644)
	if err != nil {
		return err
	}
	return nil
}
