package templater

import (
	"embed"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

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
		column_data += fmt.Sprintf(`  ,%s::%s AS %s`, EscapePath(field.Path), field.InferredType, NormaliseKey(field.Node))
		column_data += "\n"
	}
	// strip the first comma out
	column_data = strings.Replace(column_data, ",", "", 1)
	// strip the last new line out
	column_data = column_data[0 : len(column_data)-1]

	return column_data
}

func WriteSQLModel(table Table, w io.Writer) error {
	sqlTemplate := SQLTemplate{
		Tags:    GenerateTagsSQL(table.Project, table.Name),
		Columns: GenerateColumnsSQL(table.Fields),
		Source:  GenerateSourceSQL(table.Project, table.Name),
	}
	tpl, err := template.New("template.gohtml").ParseFS(fs, "template.gohtml")
	if err != nil {
		return err
	}
	return tpl.Execute(w, sqlTemplate)
}
