package version

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	info := Get()

	// Test required fields are not empty
	if info.GitVersion == "" {
		t.Error("GitVersion should not be empty")
	}

	if info.GitCommit == "" {
		t.Error("GitCommit should not be empty")
	}

	if info.BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}

	// Test runtime information
	if info.GoVersion != runtime.Version() {
		t.Errorf("Expected GoVersion %s, got %s", runtime.Version(), info.GoVersion)
	}

	if info.Compiler != runtime.Compiler {
		t.Errorf("Expected Compiler %s, got %s", runtime.Compiler, info.Compiler)
	}

	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("Expected Platform %s, got %s", expectedPlatform, info.Platform)
	}
}

func TestGetDefaultGitVersion(t *testing.T) {
	// Test when git is available
	version := getDefaultGitVersion()
	if version == "" {
		t.Error("getDefaultGitVersion should not return empty string")
	}

	// Test fallback when git command fails
	// Temporarily modify PATH to simulate git command failure
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	version = getDefaultGitVersion()
	if version != "v0.0.0-main" {
		t.Errorf("Expected fallback version v0.0.0-main, got %s", version)
	}
}

func TestGetDefaultGitCommit(t *testing.T) {
	// Test when git is available
	commit := getDefaultGitCommit()
	if commit == "" {
		t.Error("getDefaultGitCommit should not return empty string")
	}

	// Test fallback when git command fails
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	commit = getDefaultGitCommit()
	if commit != "unknown" {
		t.Errorf("Expected fallback commit 'unknown', got %s", commit)
	}
}

func TestVersionInfoFields(t *testing.T) {
	info := Get()

	// Test GitVersion format
	if !strings.HasPrefix(info.GitVersion, "v") && info.GitVersion != "v0.0.0-main" {
		t.Errorf("GitVersion should start with 'v' or be 'v0.0.0-main', got %s", info.GitVersion)
	}

	// Test GitCommit format (should be a short hash or 'unknown')
	if len(info.GitCommit) > 8 && info.GitCommit != "unknown" {
		t.Errorf("GitCommit should be a short hash or 'unknown', got %s", info.GitCommit)
	}

	// Test BuildDate format
	if _, err := time.Parse(time.RFC1123, info.BuildDate); err != nil {
		t.Errorf("BuildDate should be in RFC1123 format, got %s", info.BuildDate)
	}
}
