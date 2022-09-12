package templater_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mr-joshcrane/templater"
)

func TestGenerateTemplateGivenLowercaseTableOrProjectCorrectlyUppercases(t *testing.T) {
	t.Parallel()
	got := templater.GenerateTemplate("fixtures/data1.json", "project", "table")
	want := `{{ config(tags=['PROJECT', 'TABLE']) }}

SELECT
	"V":id::STRING AS ID,
	"V":name::STRING AS NAME
FROM
	{{ source('PROJECT', 'TABLE') }}
`
	if want != got {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGenerateTemplateGivenUnstructuredDataReturnsValidTemplate(t *testing.T) {
	t.Parallel()
	got := templater.GenerateTemplate("fixtures/data.json", "PROJECT", "TABLE")
	want := `{{ config(tags=['PROJECT', 'TABLE']) }}

SELECT
	"V":id::STRING AS ID,
	"V":orderindex::INTEGER AS ORDERINDEX,
	"V":floatedOrder::FLOAT AS FLOATEDORDER,
	"V":status::OBJECT AS STATUS,
	"V":assignee::VARCHAR AS ASSIGNEE,
	"V":task_count::ARRAY AS TASK_COUNT,
	"V":archived::BOOLEAN AS ARCHIVED
FROM
	{{ source('PROJECT', 'TABLE') }}
`
	if want != got {
		t.Fatal(cmp.Diff(want, got))
	}
}
