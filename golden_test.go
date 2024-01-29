package pql

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var recordGoldens = flag.Bool("record", false, "output golden files")

func TestGoldens(t *testing.T) {
	dir := filepath.Join("testdata", "Goldens")
	dirListing, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range dirListing {
		testName := entry.Name()
		if !entry.IsDir() || strings.HasPrefix(testName, ".") || strings.HasPrefix(testName, "_") {
			continue
		}

		t.Run(testName, func(t *testing.T) {
			testDir := filepath.Join(dir, testName)
			input, err := os.ReadFile(filepath.Join(testDir, "input.pql"))
			if err != nil {
				t.Fatal(err)
			}

			got, err := Compile(string(input))
			if err != nil {
				t.Error("Compile(...):", err)
			}

			outputPath := filepath.Join(testDir, "output.sql")
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
