// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

var raw = []byte(`- table: users
  rows:
    - _id: smith
      name: smith
      age: 10
    - _id: john
      name: john
      age: 20
- table: accounts
  rows:
    - user_id: $users.smith.id
      type: Facebook
    - user_id: $users.smith.id
      type: Google
- table: books
  rows:
    - author_id: $authors.smith
      name: Amazing Book
      book_category_id: $book_categories.mystery
- table: authors
  rows:
    - _id: smith
      user_id: $users.smith.id
      penname: smith
- table: book_categories
  rows:
    - _id: mystery
      name: mystery
`)

type Record struct {
	Table string              `json:"table"`
	Rows  []map[string]string `json:"rows"`
}

type Dep struct {
	value string
	deps  []string
}

func main() {
	var records []Record
	err := yaml.Unmarshal(raw, &records)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	depsByTable := make(map[string]map[string]bool)
	// Find all records with dependencies first, and try to resolve them first.
	for _, r := range records {
		for _, row := range r.Rows {
			for _, v := range row {
				if strings.HasPrefix(v, "$") {
					paths := strings.Split(v[1:], ".")
					dependsOnTable := paths[0]
					if _, ok := depsByTable[r.Table]; !ok {
						depsByTable[r.Table] = make(map[string]bool)
					}
					depsByTable[r.Table][dependsOnTable] = true
				}
			}
		}
	}
	var orderedDeps []string
	var traverse func(string)
	traverse = func(table string) {
		if v, ok := depsByTable[table]; ok {
			for k := range v {
				traverse(k)
			}
		}
		var exists bool
		for _, v := range orderedDeps {
			if v == table {
				exists = true
				break
			}
		}
		if exists {
			return
		}
		orderedDeps = append(orderedDeps, table)
	}
	for k := range depsByTable {
		traverse(k)
	}
	fmt.Println(orderedDeps)
}
