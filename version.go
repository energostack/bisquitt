package bisquitt

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var version string

func Version() string {
	return strings.TrimSpace(version)
}
