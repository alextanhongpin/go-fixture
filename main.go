// You can edit this code!
// Click here and start typing.
package fixture

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"sort"
	"strings"
)

var (
	ErrDuplicateAlias = errors.New("duplicate alias")
)

type FS interface {
	Open(name string) (fs.File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

type Record struct {
	Table string                   `json:"table"`
	Alias *string                  `json:"alias"`
	Rows  []map[string]interface{} `json:"rows"`
}

type Dep struct {
	value string
	deps  []string
}

func Parse(raw []byte, unmarshalFn func([]byte, any) error) string {
	var records []Record
	err := unmarshalFn(raw, &records)
	if err != nil {
		panic(err)
	}

	hasDependenciesByTable := make(map[string]bool)
	depsByTable := make(map[string]map[string]bool)
	rowsByTable := make(map[string][]map[string]interface{})
	tableByAlias := make(map[string]string)

	// NOTE: The insert statement is actually done in reverse order.
	for _, r := range records {
		reverse(r.Rows)
	}

	// Resolve all alias first.
	for _, r := range records {
		if alias := r.Alias; alias != nil {
			if t, exists := tableByAlias[*alias]; exists {
				panic(fmt.Errorf("%w: table %q and %q using the alias %q", ErrDuplicateAlias, t, r.Table, *alias))
			}

			tableByAlias[*alias] = r.Table
		}
	}

	// Find all records with dependencies first, and try to resolve them first.
	for _, r := range records {
		if _, ok := hasDependenciesByTable[r.Table]; !ok {
			hasDependenciesByTable[r.Table] = false
		}
		if _, ok := rowsByTable[r.Table]; !ok {
			rowsByTable[r.Table] = make([]map[string]interface{}, 0)
		}
		rowsByTable[r.Table] = append(rowsByTable[r.Table], r.Rows...)
		for _, row := range r.Rows {
			for _, v := range row {
				s := fmt.Sprint(v)
				if strings.HasPrefix(s, "$") {
					paths := strings.Split(s[2:], ".")
					tableOrAlias := paths[0]
					dependsOnTable := paths[0]
					if tbl, ok := tableByAlias[tableOrAlias]; ok {
						dependsOnTable = tbl
					}

					hasDependenciesByTable[dependsOnTable] = true
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

	remaining := make(map[string]bool)
	for k, v := range hasDependenciesByTable {
		remaining[k] = v
	}

	for _, v := range orderedDeps {
		delete(remaining, v)
	}

	for k := range remaining {
		orderedDeps = append(orderedDeps, k)
	}

	stmts := make([]string, 0)
	for _, v := range orderedDeps {
		rows := rowsByTable[v]
		for i, row := range rows {
			var keys, values []string
			for k := range row {
				if k == "_id" {
					continue
				}
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				v := row[k]
				s := fmt.Sprint(v)
				// $.products.coke.id
				if strings.HasPrefix(s, "$") {
					parts := strings.Split(s[2:], ".") // Output: [products, coke, id]
					col := parts[len(parts)-1]         // Output: id

					tbl := parts[:len(parts)-1] // Output: [products, coke]
					if v, ok := tableByAlias[tbl[0]]; ok {
						tbl[0] = v
					}
					stmt := fmt.Sprintf(`(SELECT %s FROM %s)`, col, strings.Join(tbl, "_"))
					values = append(values, stmt)
				} else {

					switch v.(type) {
					case nil:
						values = append(values, "NULL")
					case string:
						// Is a function, e.g. fn:now().
						if strings.Contains(s, "fn:") && strings.Contains(s, "(") && strings.Contains(s, ")") {
							s = strings.ReplaceAll(s, "fn:", "")
							values = append(values, s)
							continue
						}

						// Can only be string.
						values = append(values, fmt.Sprintf(`'%s'`, v))
					default:
						values = append(values, s)
					}
				}
			}
			tbl := v
			id, ok := row["_id"]
			if ok {
				tbl = fmt.Sprintf("%s_%s", tbl, id)
			} else {
				tbl = fmt.Sprintf("%s_%d", tbl, i)
			}

			stmts = append(stmts, fmt.Sprintf(`  %s AS (INSERT INTO %s(%s) VALUES (%s) RETURNING *)`, tbl, v, strings.Join(keys, ", "), strings.Join(values, ", ")))
		}
	}

	return strings.Join([]string{
		"WITH",
		strings.Join(stmts, ",\n"),
		"SELECT;",
	}, "\n")
}

func ParseFS(fs FS, dir string, unmarshalFn func([]byte, any) error) []string {
	dirs, err := fs.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	var result []string
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		raw, err := fs.ReadFile(dir.Name())
		if err != nil {
			panic(err)
		}
		result = append(result, Parse(raw, unmarshalFn))
	}
	return result
}

func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
