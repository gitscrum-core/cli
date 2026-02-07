// GitScrum CLI - Entry Point
// Minimal main.go following GitHub CLI pattern
package main

import (
	"os"

	"github.com/gitscrum-core/cli/pkg/cmd/root"
)

// Version info - set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	code := root.Execute(version, commit, date)
	os.Exit(code)
}
