[![Go Reference](https://pkg.go.dev/badge/github.com/mr-joshcrane/templater.svg)](https://pkg.go.dev/github.com/mr-joshcrane/templater)[![License: GPL-2.0](https://img.shields.io/badge/Licence-GPL-2)](https://opensource.org/licenses/GPL-2.0)[![Go Report Card](https://goreportcard.com/badge/github.com/mr-joshcrane/templater)](https://goreportcard.com/report/github.com/mr-joshcrane/templater)

# templater

**Quick Install**
```bash
go install github.com/mr-joshcrane/templater/cmd/templater@latest
```
---

Templater generates DBT Models from raw CSV files.

These Models have type inference, key normalisation, and generate most of the boilerplate code for you.


It can also spare you the drudgery of "unpacking" variant payloads in your staging SQL models.  

Usage: 

`templater [FIELDS_TO_UNPACK ...]`

FIELDS_TO_UNPACK is an optional indications of which fields are JSON objects, capable of further unpacking.

1. Start with a raw CSV
![Start with a CSV](docs/csv.png)
___
2. Run templater from the directory containing your collection of CSV files, optionally specifying names of any JSON blobs to unpack 

    `$ templater Career`
___
3. Get a nice DBT model and project config files in return!
![Start with a CSV](docs/sql_models.png)
