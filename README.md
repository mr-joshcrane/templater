[![Go Reference](https://pkg.go.dev/badge/github.com/mr-joshcrane/templater.svg)](https://pkg.go.dev/github.com/mr-joshcrane/templater)[![License: GPL-2.0](https://img.shields.io/badge/Licence-GPL-2)](https://opensource.org/licenses/GPL-2.0)[![Go Report Card](https://goreportcard.com/badge/github.com/mr-joshcrane/templater)](https://goreportcard.com/report/github.com/mr-joshcrane/templater)

# templater

**Quick Install**
```bash
go install github.com/mr-joshcrane/templater/cmd/templater@latest
```

**Usage**

```bash
$ templater [FIELDS_TO_UNPACK ...]
```

FIELDS_TO_UNPACK is an optional indications of which fields are JSON objects, capable of further unpacking.

---
## Why would you use templater?
Data Engineering will often require taking some raw, untyped and unsanitised data and running it through a series of preliminary transformations before it can be presented in its final format. 

Templater was created with DBT as the transformation tool in mind, and is designed to generate a complete DBT project from a set of CSV files. These CSVs are typically exported Snowflake tables, but any CSVs can be used.

Theoretically templater can be used with any data warehouse solution, but it has only been tested with Snowflake.

## Data doesn't always play nice
Sometimes you'll have some untyped data that needs to be typed. This usually isn't difficult, but is very tedious and labour intensive. 

Sometimes you will have column names that contain special characters that are not allowed in Snowflake (without special handling).

Sometimes you will have nested JSON data that needs to be flattened before it can be used for any real business application.

Templater can be used to allieviate some of the burden here!

---
Imagine we had a table in Snowflake called `ENERGY` and we dumped it out to a CSV file as `ENERGY.csv`:

<table>
  <tr>
    <th>sourceAsMetric</th><th>target(as %)</th><th>value(per million tonnes)</th><th>statistics</th>
  </tr>
  <tr>
    <td>Bio-conversion(natural)</td><td>Liquid</td><td>124.729</td><td>"{  ""attributes"": {    ""on_hand"": false,    ""available_in"": ""days"",    ""componentDemands"": [99,2.5,"{5,2,5}"]}}"</td>
  </tr>
  <tr>
    <td>BiofuelImports<td>Solid</td><td>35</td><td>"{  ""attributes"": {    ""on_hand"": true,    ""available_in"": ""minutes"",    ""componentDemands"": [1,0,"{0.1,9,0}"]}}"</td>
  </tr>
  <tr>
    <td>coal imports(55%)</td><td>Coal</td><td>null</td><td>"{  ""attributes"": {    ""on_hand"": true,    ""available_in"": ""hours"",    ""componentDemands"": [1,0,"{0.1,00.1,NA}"]}}"</td>
  </tr>
</table>


Gross! As we can see, the format of this data leaves  much to be desired. We could munge it with DBT, but that would be a non-trivial ammount effort!



We can use templater to generate a DBT project that will transform this data into something more useful.

Lets call templater, and indicate we want to unpack the CSV field called `statistics` which seems to be a JSON field.

```bash
$ templater statistics
```
Templater will generate the core of a new DBT project, including our transformation below.

*output/transform/TRANS01_ENERGY.sql*
```sql output/transform/TRANS01_ENERGY.sql
{{ config(materialized='table') }}
SELECT
  "statistics":"attributes"."available_in"::STRING AS ATTRIBUTES__AVAILABLE_IN
  ,"statistics":"attributes"."componentDemands"::ARRAY AS ATTRIBUTES__COMPONENT_DEMANDS
  ,"statistics":"attributes"."on_hand"::BOOLEAN AS ATTRIBUTES__ON_HAND
  ,"sourceAsMetric"::STRING AS SOURCE_AS_METRIC
  ,"target(idealised)"::STRING AS TARGET_IDEALISED
  ,"value(per million tonnes)"::STRING AS VALUE_PER_MILLION_TONNES
FROM
  {{ source('TEMPLATER', 'ENERGY') }}
```

We can see that templater has:

1. Generated our transformation SQL
2. Tried to infer Snowflake Types for each column
3. Normalised the column names to be Snowflake friendly

It has also generated some suggested DBT models that you can tweak to your liking in *output/transform/_models_schema.yml*, which will go a long way when you're trying to generate some `dbt docs`.

```yaml
version: 2
models:
  - name: TRANS01_ENERGY
    columns:
      - name: ATTRIBUTES__AVAILABLE_IN
      - name: ATTRIBUTES__COMPONENT_DEMANDS
      - name: ATTRIBUTES__ON_HAND
      - name: SOURCE_AS_METRIC
      - name: TARGET_IDEALISED
      - name: VALUE_PER_MILLION_TONNES

```
