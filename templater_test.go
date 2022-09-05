package templater_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mr-joshcrane/templater"
)

func TestGenerateTemplate(t *testing.T) {
	t.Parallel()
	//"fixtures/data.json"

	got := templater.GenerateTemplate("fixtures/data.json", "clickup", "tasks")

	want := `{{ config(tags=['CLICKUP', 'TASKS']) }}

SELECT
	"V":id::STRING AS ID,
	"V":name::STRING AS NAME,
	"V":orderindex::INTEGER AS ORDERINDEX,
	"V":content::STRING AS CONTENT,
	"V":status::OBJECT AS STATUS,
	"V":priority::OBJECT AS PRIORITY,
	"V":assignee::VARCHAR AS ASSIGNEE,
	"V":task_count::ARRAY AS TASK_COUNT,
	"V":due_date::STRING AS DUE_DATE,
	"V":start_date::VARCHAR AS START_DATE,
	"V":folder::OBJECT AS FOLDER,
	"V":space::OBJECT AS SPACE,
	"V":archived::BOOLEAN AS ARCHIVED,
	"V":override_statuses::BOOLEAN AS OVERRIDE_STATUSES,
	"V":permission_level::STRING AS PERMISSION_LEVEL,
FROM
	{{ source('CLICKUP', 'TASKS') }}
`
	if want != got {
		t.Fatal(cmp.Diff(want, got))
	}
}
