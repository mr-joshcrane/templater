package templater

import (
	"fmt"
	"sort"
)

type Test struct {
	Test string `yaml:"test, omitempty"`
}
type Column struct {
	Name        string   `yaml:"name"`
	Description *string   `yaml:"description, omitempty"`
	Tests       []string `yaml:"tests, omitempty"`
}
type Sources struct {
	Version int `yaml:"version"`
	Sources []Source `yaml:"sources"`
}

type Source struct {
	Name   string   `yaml:"name"`
	Schema string   `yaml:"schema"`
	Tables []Column `yaml:"tables, omitempty"`
}

type Models struct {
	Version int     `yaml:"version"`
	Models  []Model `yaml:"models"`
}

type Model struct {
	Name        string   `yaml:"name"`
	Description *string   `yaml:"description, omitempty"`
	Tests       []Test   `yaml:"tests, omitempty"`
	Columns     []Column `yaml:"columns"`
}

func GenerateModel(tables []Table) Models {
	var models []Model
	for _, v := range tables {
		m := Model{}
		m.Name = v.Name
		for k := range v.TypeMap {
			k = formatKey(k)
			col := Column{
				Name: k,
			}
			m.Columns = append(m.Columns, col)
			sort.Slice(m.Columns, func(i, j int) bool {
				return m.Columns[i].Name < m.Columns[j].Name
			})
		}
		models = append(models, m)
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
	return Models{
		Version: 2,
		Models:  models,
	}
}

func (m *Models) AddDescriptions() *Models {
	for model := range m.Models {
		modelDescription := fmt.Sprintf("TODO: Description for MODEL, %s", m.Models[model].Name)
		m.Models[model].Description = &modelDescription
		for column := range m.Models[model].Columns {
			columnDescription := fmt.Sprintf("TODO: Description for COLUMN, %s", m.Models[model].Columns[column].Name)
			m.Models[model].Columns[column].Description = &columnDescription
		}
	}
	return m
}

func generateSources(tables []Table, projectName string) Sources {
	var source Source

	source.Name = projectName
	source.Schema = "STAGING"
	for _, column := range tables {
		columnDescription := fmt.Sprintf("TODO: Description for TABLE, %s", column.Name)
		t := Column{
			Name:        column.Name,
			Description: &columnDescription,
		}
		source.Tables = append(source.Tables, t)
		sort.Slice(source.Tables, func(i, j int) bool {
			return source.Tables[i].Name < source.Tables[j].Name
		})
	}
	return Sources{
		Version: 2,
		Sources: []Source{source},
	}
}
