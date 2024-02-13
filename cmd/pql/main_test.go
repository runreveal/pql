package main

import (
	"context"
	"strings"
	"testing"

	"github.com/runreveal/pql"
)

func TestRun(t *testing.T) {
	const inputStatement = "StormEvents"
	outputStatement, err := pql.Compile(inputStatement)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		input  string
		output string
		fail   bool
	}{
		{
			name: "Empty",
			fail: false,
		},
		{
			name:   "WhitespaceOnly",
			input:  " \t \n\n\n",
			output: "",
		},
		{
			name:   "CommentOnly",
			input:  "// This is a comment.\n\n",
			output: "",
		},
		{
			name:   "Statement",
			input:  inputStatement + "\n",
			output: outputStatement + "\n\n",
		},
		{
			name:  "BadStatement",
			input: "!",
			fail:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			gotOutput := new(strings.Builder)
			gotError := run(ctx, gotOutput, strings.NewReader(test.input), func(error) {})

			if got := gotOutput.String(); got != test.output {
				t.Errorf("output = %q; want %q", got, test.output)
			}
			if (gotError != nil) && !test.fail {
				t.Errorf("unexpected error %v", gotError)
			}
			if gotError == nil && test.fail {
				t.Error("did not return an error")
			}
		})
	}
}
