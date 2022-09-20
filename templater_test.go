package templater_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mr-joshcrane/templater"
)

func TestGenerateTemplateGivenLowercaseTableOrProjectCorrectlyUppercases(t *testing.T) {
	t.Parallel()
	got, err := templater.GenerateTemplate("fixtures/data1.json", "project", "table")
	if err != nil {
		t.Fatalf("wasn't expecting error, but got %v", err)
	}
	want := `{{ config(tags=['PROJECT', 'TABLE']) }}

SELECT
	"V":id::STRING AS ID
	,"V":name::STRING AS NAME
FROM
	{{ source('PROJECT', 'TABLE') }}
`
	if want != got {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGenerateTemplateGivenUnstructuredDataReturnsValidTemplate(t *testing.T) {
	t.Parallel()
	got, err := templater.GenerateTemplate("fixtures/data.json", "PROJECT", "TABLE")
	if err != nil {
		t.Fatalf("wasn't expecting error, but got %v", err)
	}
	want := `{{ config(tags=['PROJECT', 'TABLE']) }}

SELECT
	"V":id::STRING AS ID
	,"V":orderindex::INTEGER AS ORDERINDEX
	,"V":floatedOrder::FLOAT AS FLOATEDORDER
	,"V":status::OBJECT AS STATUS
	,"V":assignee::VARCHAR AS ASSIGNEE
	,"V":task_count::ARRAY AS TASK_COUNT
	,"V":archived::BOOLEAN AS ARCHIVED
FROM
	{{ source('PROJECT', 'TABLE') }}
`
	fmt.Println(want)
	fmt.Println(got)
	if want != got {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGenerateOnEmptyJSONShouldReturnEmptyJSONError(t *testing.T) {
	t.Parallel()
	_, err := templater.GenerateTemplate("fixtures/data2.json", "PROJECT", "TABLE")
	if err.Error() != "empty JSON" {
		t.Fatal(err.Error())
	}
}

func TestGenerateOnInvalidJSONShouldReturnInvalidJSONError(t *testing.T) {
	t.Parallel()
	_, err := templater.GenerateTemplate("fixtures/data3.json", "PROJECT", "TABLE")
	if err.Error() != "unable to convert json to cue" {
		t.Fatal(err.Error())
	}
}
