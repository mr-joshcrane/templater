package templater_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mr-joshcrane/templater"
)

func must(bs []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return bs
}

func TestGenerateTemplateGivenLowercaseTableOrProjectCorrectlyUppercases(t *testing.T) {
	t.Parallel()
	contents := must(os.ReadFile("fixtures/data1.json"))
	got, err := templater.GenerateTemplate(contents, "project", "table")
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

func TestGenerateTemplateGivenCsv(t *testing.T) {
	t.Parallel()
	csvContents := must(os.ReadFile("fixtures/data.csv"))
	contents, err := templater.CsvToJson(csvContents)
	if err != nil {
		t.Fatalf("wasn't expecting error, but got %v", err)
	}
	got, err := templater.GenerateTemplate(contents, "project", "table")
	if err != nil {
		t.Fatalf("wasn't expecting error, but got %v", err)
	}
	want := `{{ config(tags=['PROJECT', 'TABLE']) }}

SELECT
	"V":id::STRING AS ID
	,"V":orderindex::STRING AS ORDERINDEX
	,"V":value::STRING AS VALUE
	,"V":quantity::STRING AS QUANTITY
	,"V":fieldbuf::STRING AS FIELDBUF
FROM
	{{ source('PROJECT', 'TABLE') }}
`
	if want != got {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGenerateTemplateGivenUnstructuredDataReturnsValidTemplate(t *testing.T) {
	t.Parallel()
	contents := must(os.ReadFile("fixtures/data.json"))
	got, err := templater.GenerateTemplate(contents, "PROJECT", "TABLE")
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
	if want != got {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestGenerateOnEmptyJSONShouldReturnEmptyJSONError(t *testing.T) {
	t.Parallel()
	contents := must(os.ReadFile("fixtures/data2.json"))
	_, err := templater.GenerateTemplate(contents, "PROJECT", "TABLE")
	if err.Error() != "empty JSON" {
		t.Fatal(err.Error())
	}
}

func TestGenerateOnInvalidJSONShouldReturnInvalidJSONError(t *testing.T) {
	t.Parallel()
	contents := must(os.ReadFile("fixtures/data3.json"))
	_, err := templater.GenerateTemplate(contents, "PROJECT", "TABLE")
	if err.Error() != "unable to convert json to cue" {
		t.Fatal(err.Error())
	}
}
