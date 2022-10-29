package templater

import (
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
	filename := fmt.Sprintf("output/%s.sql", table.Name)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	sqlTemplate := SQLTemplate{
		Tags:    GenerateTagsSQL(table.Project, table.Name),
		Columns: GenerateColumnsSQL(table.Fields),
		Source:  GenerateSourceSQL(table.Project, table.Name),
	}
	tpl, err := template.New("template.gohtml").ParseFS(fs, "template.gohtml")
	if err != nil {
		return err
	}
	return tpl.Execute(file, sqlTemplate)
}

func WriteProperties(c *cue.Context, models Models, sources Sources) error {
	err := WriteProperty("transform_schema.yml", c, models)
	if err != nil {
		return err
	}
	err = WriteProperty("public_schema.yml", c, *models.AddDescriptions())
	if err != nil {
		return err
	}
	err = WriteProperty("source_schema.yml", c, sources)
	if err != nil {
		return err
	}
	return nil
}

func WriteProperty[T Sources | Models](path string, c *cue.Context, t T) error {
	encoded, err := yaml.Encode(c.Encode(t))
	if err != nil {
		return err
	}
	path = fmt.Sprintf("output/%s", path)
	err = os.WriteFile(path, encoded, 0644)
	if err != nil {
		return err
	}
	return nil
}
