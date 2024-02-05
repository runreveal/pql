package pql

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClickhouseLocal(t *testing.T) {
	clickhouseExe, err := exec.LookPath("clickhouse")
	if err != nil {
		t.Skipf("Skipping: clickhouse not found: %v", err)
	}

	tests, err := findGoldenTests()
	if err != nil {
		t.Fatal(err)
	}
	tables, err := findLocalTables(filepath.Join("testdata", "Tables"))
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		wantCSV, wantCSVError := os.ReadFile(filepath.Join(test.dir, "output.csv"))
		if errors.Is(wantCSVError, os.ErrNotExist) {
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.Skipf("'skip' file present in %s; skipping...", test.dir)
			}
			if wantCSVError != nil {
				t.Fatal("Could not read expected output:", wantCSVError)
			}

			pqlInput, err := test.input()
			if err != nil {
				t.Fatal(err)
			}
			query, err := Compile(pqlInput)
			if err != nil {
				t.Fatal("Compile:", err)
			}

			var args []string
			args = append(args, "local", "--format", "CSV")
			sb := new(strings.Builder)
			for _, tab := range tables {
				sb.Reset()
				quoteSQLString(sb, tab.filename)
				args = append(args, "--query", fmt.Sprintf("CREATE TABLE \"%s\" AS file(%s);", tab.name, sb))
			}
			args = append(args, "--query", query)

			c := exec.Command(clickhouseExe, args...)
			c.Dir = test.dir
			gotCSV := new(bytes.Buffer)
			c.Stdout = gotCSV
			stderr := new(bytes.Buffer)
			c.Stderr = stderr
			runError := c.Run()
			if stderr.Len() > 0 {
				t.Logf("clickhouse local stderr:\n%s", stderr)
			}
			if runError != nil {
				t.Fatal("clickhouse local failed:", runError)
			}

			got, err := csv.NewReader(gotCSV).ReadAll()
			if err != nil {
				t.Fatal(err)
			}
			want, err := csv.NewReader(bytes.NewReader(wantCSV)).ReadAll()
			if err != nil {
				t.Fatal(err)
			}

			if test.unordered {
				sort.Slice(got, func(i, j int) bool {
					return isRowLess(got[i], got[j])
				})
				sort.Slice(want, func(i, j int) bool {
					return isRowLess(want[i], want[j])
				})
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("query results (-want +got):\n%s", diff)
			}
		})
	}
}

func isRowLess(row1, row2 []string) bool {
	for i, n := 0, min(len(row1), len(row2)); i < n; i++ {
		if row1[i] < row2[i] {
			return true
		}
		if row1[i] > row2[i] {
			return false
		}
	}
	return len(row1) < len(row2)
}

type localTable struct {
	name     string
	filename string
}

// findLocalTables finds all CSV files in a directory that represent tables.
func findLocalTables(dir string) ([]localTable, error) {
	var err error
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("find local tables: %v", err)
	}
	listing, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("find local tables: %v", err)
	}

	var result []localTable
	for _, entry := range listing {
		filename := entry.Name()
		baseName, isCSV := strings.CutSuffix(filename, ".csv")
		if entry.Type().IsRegular() && isCSV && !shouldIgnoreFilename(filename) {
			result = append(result, localTable{
				name:     baseName,
				filename: filepath.Join(dir, filename),
			})
		}
	}
	return result, nil
}
