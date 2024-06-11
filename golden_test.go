// Copyright 2024 RunReveal Inc.
// SPDX-License-Identifier: Apache-2.0

package pql

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tailscale/hujson"
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
			compileOptions, _, err := test.options()
			if err != nil {
				t.Fatal(err)
			}

			got, err := compileOptions.Compile(input)
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
	name      string
	dir       string
	skip      bool
	unordered bool
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
		if _, err := os.Stat(filepath.Join(test.dir, "unordered")); err == nil {
			test.unordered = true
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("find golden tests: check for unordered: %v", err)
		}
		result = append(result, test)
	}
	return result, nil
}

func (test goldenTest) input() (string, error) {
	input, err := os.ReadFile(filepath.Join(test.dir, "input.pql"))
	return string(input), err
}

type testOptions struct {
	parameterValues map[string]string
}

func (test goldenTest) options() (*CompileOptions, *testOptions, error) {
	type testParameter struct {
		Value string `json:"value"`
		SQL   string `json:"clickhouse"`
	}

	path := filepath.Join(test.dir, "options.jwcc")
	input, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, new(testOptions), nil
	}
	if err != nil {
		return nil, nil, err
	}
	input, err = hujson.Standardize(input)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %v", path, err)
	}
	var parsed struct {
		Parameters map[string]testParameter `json:"parameters"`
	}
	if err := json.Unmarshal(input, &parsed); err != nil {
		return nil, nil, fmt.Errorf("parse %s: %v", path, err)
	}
	opts := &CompileOptions{
		Parameters: make(map[string]string, len(parsed.Parameters)),
	}
	testOpts := &testOptions{
		parameterValues: make(map[string]string, len(parsed.Parameters)),
	}
	for name, p := range parsed.Parameters {
		opts.Parameters[name] = p.SQL
		testOpts.parameterValues[name] = p.Value
	}
	return opts, testOpts, nil
}

func shouldIgnoreFilename(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}
