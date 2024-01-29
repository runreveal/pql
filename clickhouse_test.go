package pql

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const outputCSVFilename = "output.csv"

func TestClickhouseLocal(t *testing.T) {
	clickhouseExe, err := exec.LookPath("clickhouse")
	if err != nil {
		t.Skipf("Skipping: clickhouse not found: %v", err)
	}

	tests, err := findGoldenTests()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		wantCSV, wantCSVError := os.ReadFile(filepath.Join(test.dir, outputCSVFilename))
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
			tables, err := findLocalTables(test.dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, tab := range tables {
				args = append(args, "--query", fmt.Sprintf("CREATE TABLE %s AS file('%s');", tab.name, tab.filename))
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
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("query results (-want +got):\n%s", diff)
			}
		})
	}
}

type localTable struct {
	name     string
	filename string
}

// findLocalTables finds all CSV files in a golden test directory that represent tables.
func findLocalTables(dir string) ([]localTable, error) {
	listing, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("find local tables: %v", err)
	}
	var result []localTable
	for _, entry := range listing {
		filename := entry.Name()
		if filename == outputCSVFilename {
			continue
		}
		baseName, isCSV := strings.CutSuffix(filename, ".csv")
		if entry.Type().IsRegular() && isCSV && !shouldIgnoreFilename(filename) {
			result = append(result, localTable{
				name:     baseName,
				filename: filename,
			})
		}
	}
	return result, nil
}
