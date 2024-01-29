package pql

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var recordGoldens = flag.Bool("record", false, "output golden files")

func TestGoldens(t *testing.T) {
	tests, err := findGoldenTests()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.Skipf("'skip' file present in %s; skipping...", test.dir)
			}

			input, err := test.input()
			if err != nil {
				t.Fatal(err)
			}

			got, err := Compile(input)
			if err != nil {
				t.Error("Compile(...):", err)
			}

			outputPath := filepath.Join(test.dir, "output.sql")
			if *recordGoldens {
				// For easier editing, ensure there is a trailing newline.
				if got != "" && !strings.HasSuffix(got, "\n") {
					got += "\n"
				}

				if err := os.WriteFile(outputPath, []byte(got), 0o666); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			// Strip trailing newlines for comparison.
			// Makes it easier to hand-edit goldens when editors place trailing newlines.
			got = strings.TrimRight(got, "\n")
			want = bytes.TrimRight(want, "\n")
			if diff := cmp.Diff(string(want), got); diff != "" {
				t.Errorf("output (-want +got):\n%s", diff)
			}
		})
	}
}

type goldenTest struct {
	name string
	dir  string
	skip bool
}

func findGoldenTests() ([]goldenTest, error) {
	root := filepath.Join("testdata", "Goldens")
	rootListing, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("find golden tests: %v", err)
	}

	var result []goldenTest
	for _, entry := range rootListing {
		fileName := entry.Name()
		if !entry.IsDir() || shouldIgnoreFilename(fileName) {
			continue
		}
		test := goldenTest{
			name: fileName,
			dir:  filepath.Join(root, fileName),
		}
		if _, err := os.Stat(filepath.Join(test.dir, "skip")); err == nil {
			test.skip = true
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("find golden tests: check for skip: %v", err)
		}
		result = append(result, test)
	}
	return result, nil
}

func (test goldenTest) input() (string, error) {
	input, err := os.ReadFile(filepath.Join(test.dir, "input.pql"))
	return string(input), err
}

func shouldIgnoreFilename(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}
