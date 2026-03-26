package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION.txt
var versionFile string

// Version returns the current version of the application
func Version() string {
	return strings.TrimSpace(versionFile)
}
