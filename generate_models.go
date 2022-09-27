package templater

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
)

type SourceTable struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type Source struct {
	Name   string        `yaml:"name"`
	Schema string        `yaml:"schema"`
	Tables []SourceTable `yaml:"tables"`
}

type Sources struct {
	Version int      `yaml:"version"`
	Sources []Source `yaml:"sources"`
}

type TransformColumn struct {
	Name  string   `yaml:"name"`
	Tests []string `yaml:"tests"`
}

type TransformModel struct {
	Name    string            `yaml:"name"`
	Tests   []string          `yaml:"tests"`
	Columns []TransformColumn `yaml:"columns"`
}

type TransformModels struct {
	Version int              `yaml:"version"`
	Models  []TransformModel `yaml:"models"`
}

type PublicColumn struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tests       []string `yaml:"tests"`
}

type PublicModel struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Tests       RowCountTest   `yaml:"tests"`
	Columns     []PublicColumn `yaml:"columns"`
}

type PublicModels struct {
	Version int           `yaml:"version"`
	Models  []PublicModel `yaml:"models"`
}

type CompareModel struct {
	CompareModel string `yaml:"compare_model"`
}

type RowCountTest struct {
	RowCountTest CompareModel `yaml:"dbt_utils.equal_rowcount"`
}

func generateTransform(c *cue.Context, metadata Metadata) (string, error) {
	var z TransformModels
	z.Version = 2
	z.Models = []TransformModel{}

	for _, v := range metadata.Tables {
		m := TransformModel{}
		m.Name = v.TableName
		m.Tests = []string{"test"}
		for k := range v.TypeMap {
			k = formatKey(k)
			col := TransformColumn{
				Name:  k,
				Tests: []string{"not_null"},
			}
			m.Columns = append(m.Columns, col)
		}

		z.Models = append(z.Models, m)
	}

	cModel := c.Encode(z)
	yaml, err := yaml.Encode(cModel)
	if err != nil {
		panic(err)
	}
	return string(yaml), err
}

func generatePublic(c *cue.Context, metadata Metadata) (string, error) {
	var z PublicModels
	z.Version = 2
	z.Models = []PublicModel{}
	for _, v := range metadata.Tables {
		m := PublicModel{}
		m.Name = v.TableName
		m.Description = fmt.Sprintf("TODO: Description for %s", m.Name)
		for k := range v.TypeMap {
			k = formatKey(k)
			col := PublicColumn{
				Name:        k,
				Description: fmt.Sprintf("TODO: Description for %s", k),
				Tests:       []string{"not_null"},
			}
			m.Columns = append(m.Columns, col)
		}
		z.Models = append(z.Models, m)
	}
	cModel := c.Encode(z)
	yaml, err := yaml.Encode(cModel)
	if err != nil {
		panic(err)
	}
	return string(yaml), err
}

func generateSources(c *cue.Context, tables map[string]Table, projectName string) (string, error) {
	var z Sources
	z.Version = 2
	s := Source{}

	s.Name = projectName
	s.Schema = "STAGING"
	for k := range tables {
		t := SourceTable{
			Name:        k,
			Description: fmt.Sprintf("TODO: %s DESCRIPTION", k),
		}
		s.Tables = append(s.Tables, t)
	}
	z.Sources = []Source{s}
	sModel := c.Encode(z)
	yaml, err := yaml.Encode(sModel)
	if err != nil {
		panic(err)
	}
	return string(yaml), err
}
