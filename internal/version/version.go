package version

// These values are set during build using ldflags
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// GetVersionInfo returns the version information as a string
func GetVersionInfo() string {
	return Version + " (commit: " + Commit + ", built: " + BuildDate + ")"
}
