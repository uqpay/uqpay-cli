package build

// Injected at link time via -ldflags.
var (
	Version = "dev"
	Date    = "unknown"
)
