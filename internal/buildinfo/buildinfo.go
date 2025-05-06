// Pakcage buildinfo provides build variables for ldflags.
package buildinfo

import "fmt"

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// Print writes version, date and commit message to stdout.
func Print() {
	fmt.Printf(
		"BuildVersion: %s\nBuildData: %s\nBuildCommit: %s\n",
		buildVersion, buildDate, buildCommit,
	)
}
