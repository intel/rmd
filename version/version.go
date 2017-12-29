package version

// Build information
var (
	Version   string
	Revision  string
	Branch    string
	BuildDate string
	GoVersion string
)

// Info provides the iterable version information.
var Info = map[string]string{
	"version":   Version,
	"revision":  Revision,
	"branch":    Branch,
	"buildDate": BuildDate,
	"goVersion": GoVersion,
}
