package version

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	// These variables are replaced by ldflags at build time
	gitVersion = getDefaultGitVersion()
	gitCommit  = getDefaultGitCommit()
	buildDate  = time.Now().UTC().Format(time.RFC1123) // default to current time
)

// getDefaultGitVersion returns the current git version/tag or fallback value
func getDefaultGitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "v0.0.0-main"
	}
	return strings.TrimSpace(string(out))
}

// getDefaultGitCommit returns the current git commit hash or fallback value
func getDefaultGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

type VersionInfo struct {
	GitVersion string `json:"gitVersion" yaml:"gitVersion"`
	GitCommit  string `json:"gitCommit" yaml:"gitCommit"`
	BuildDate  string `json:"buildDate" yaml:"buildDate"`
	GoVersion  string `json:"goVersion" yaml:"goVersion"`
	Compiler   string `json:"compiler" yaml:"compiler"`
	Platform   string `json:"platform" yaml:"platform"`
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() *VersionInfo {
	// These variables typically come from -ldflags settings and in
	// their absence fallback to the constants above
	return &VersionInfo{
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
