package templater

import (
	"fmt"
	"os"
	"sort"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/yaml"
)

// Test: DBT Reference: https://docs.getdbt.com/reference/resource-properties/tests.
type Test struct {
	Test string `yaml:"test, omitempty"`
}

// Column: DBT Reference: https://docs.getdbt.com/reference/resource-properties/columns.
type Column struct {
	Name        string   `yaml:"name"`
	Description *string  `yaml:"description, omitempty"`
	Tests       []string `yaml:"tests, omitempty"`
}

// Sources: DBT Reference: https://docs.getdbt.com/reference/dbt-jinja-functions/source.
type Sources struct {
	Version int      `yaml:"version"`
	Sources []Source `yaml:"sources"`
}

// Sources: DBT Reference: https://docs.getdbt.com/reference/dbt-jinja-functions/source.
type Source struct {
	Name   string   `yaml:"name"`
	Schema string   `yaml:"schema"`
	Tables []Column `yaml:"tables, omitempty"`
}

// Models: DBT Reference: https://docs.getdbt.com/docs/dbt-cloud-apis/metadata-schema-model.
type Models struct {
	Version int     `yaml:"version"`
	Models  []Model `yaml:"models"`
}

// Models: DBT Reference: https://docs.getdbt.com/docs/dbt-cloud-apis/metadata-schema-model.
type Model struct {
	Name        string   `yaml:"name"`
	Description *string  `yaml:"description, omitempty"`
	Tests       []Test   `yaml:"tests, omitempty"`
	Columns     []Column `yaml:"columns"`
}

// GenerateProject: Generate the [Models] required in _models_schema.yaml files that help define a (potentially multi-table) DBT project.
func GenerateProjectModel(tables []*Table) Models {
	var models []Model
	for _, table := range tables {
		m := Model{}
		m.Name = table.Name
		for _, field := range table.Fields {
			node := NormaliseKey(field.Node)
			col := Column{
				Name: node,
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

// addDescriptions: Add descriptions to the [Models] to help with documentation.
// Only required in the public schema, these descriptions will show up in the DBT docs.
func (m Models) addDescriptions() Models {
	models := make([]Model, len(m.Models))
	copy(models, m.Models)
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

// addPrefix: Add a prefix to the [Models] to help satisfy the name uniqueness constraints.
func (m Models) addPrefix(prefix string) Models {
	models := make([]Model, len(m.Models))
	copy(models, m.Models)
	for model := range models {
		models[model].Name = fmt.Sprintf("%s_%s", prefix, models[model].Name)
	}
	return Models{
		Version: 2,
		Models:  models,
	}
}

// generateProjectSources: Generate the [Sources] required in _source_schema.yaml files that help define a (potentially multi-table) DBT project.
// _source_schema.yaml files define DBT relations to the source tables to be transformed.
func generateProjectSources(tables []*Table, projectName string) Sources {
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

// writeProjectModels: Write the [Models] to transform/_models_schema.yml and public/_models_schema respectively.
func writeProject(c *cue.Context, models Models, sources Sources, tables []*Table) error {
	for _, table := range tables {
		err := writeTableModel(table)
		if err != nil {
			return err
		}
	}
	err := writePropertyToFile("transform/_models_schema.yml", c, models.addPrefix("TRANS01"))
	if err != nil {
		return err
	}
	err = writePropertyToFile("public/_models_schema.yml", c, models.addDescriptions())
	if err != nil {
		return err
	}
	err = writePropertyToFile("_source_schema.yml", c, sources)
	if err != nil {
		return err
	}
	return nil
}

// writePropertyToFile: takes either a [Source] or a [Model] and writes it to file
func writePropertyToFile[T Sources | Models](path string, c *cue.Context, t T) error {
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
