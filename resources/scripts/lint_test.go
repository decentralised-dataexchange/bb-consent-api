//go:build linter
// +build linter

package linter

import (
	"flag"
	"testing"

	"github.com/surullabs/lint"
	"github.com/surullabs/lint/dupl"
	"github.com/surullabs/lint/gofmt"
	"github.com/surullabs/lint/golint"
	"github.com/surullabs/lint/govet"
)

func TestLint(t *testing.T) {
	checks := lint.Group{
		gofmt.Check{},
		govet.Check{
			Args: []string{
				"--all",
				"--composites=false",
				// TODO: enable shadow; too many issues currently.
				// "--shadow",
			},
		},
		golint.Check{},

		// TODO: fix 54 issues before enabling this.
		// errcheck.Check{
		// 	Tags: "mock",
		// },

		// TODO: currently finds way too many duplicates, so disable for now.
		// dupl.Check{Threshold: 25}, // Identify duplicates

		// TODO: takes 260s to run these.
		// gosimple.Check{ // Simplification suggestions
		// 	Tags: "mock",
		// },
		// gostaticcheck.Check{ // Verify function parameters
		// 	Tags: "mock",
		// },

		// TODO: these don't support build tags at all.
		// aligncheck.Check{},    // Struct alignment issues
		// structcheck.Check{},   // Unused struct fields
		// varcheck.Check{},      // Unused global variables
	}

	packages := flag.Args()
	if len(packages) == 0 {
		t.Fatal("Pass packages to check as command line arguments")
	}

	err := checks.Check(packages...)

	err = lint.Skip(
		err,

		// Skip mocks
		lint.RegexpMatch(`mocks`, `mock_.*\.go`),

		// Skip `ContainerService` as it cannot be renamed to `Service`
		lint.RegexpMatch(`type name will be used as container\.ContainerService`),

		// Skip errcheck for command usages and deferred closes
		lint.RegexpMatch(`defer .*\.Close\(\)`),

		// Ignore all duplicates with just two instances
		dupl.SkipTwo)

	if err != nil {
		t.Fatal("lint failures:", err)
	}
}
