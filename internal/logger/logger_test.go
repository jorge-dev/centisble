package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("Text output logger", func(t *testing.T) {
		cfg := LogConfig{
			Level:      slog.LevelInfo,
			JSONOutput: false,
		}
		logger := InitLogger(cfg)

		logger.Info("test message")
		w.Close()

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r)
		if err != nil {
			t.Fatal(err)
		}

		output := buf.String()
		if !strings.Contains(output, "test message") {
			t.Errorf("Expected output to contain 'test message', got %v", output)
		}
		if strings.Contains(output, `"msg"`) {
			t.Error("Expected text output, got JSON format")
		}
	})

	// Reset stdout
	os.Stdout = oldStdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	t.Run("JSON output logger", func(t *testing.T) {
		cfg := LogConfig{
			Level:      slog.LevelDebug,
			JSONOutput: true,
		}
		logger := InitLogger(cfg)

		logger.Info("test message")
		w.Close()

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r)
		if err != nil {
			t.Fatal(err)
		}

		output := buf.String()
		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(output), &jsonMap); err != nil {
			t.Errorf("Expected valid JSON output, got error: %v", err)
		}

		if msg, ok := jsonMap["msg"]; !ok || msg != "test message" {
			t.Errorf("Expected message 'test message', got %v", msg)
		}
	})

	t.Run("Log level configuration", func(t *testing.T) {
		cfg := LogConfig{
			Level:      slog.LevelError,
			JSONOutput: false,
		}
		logger := InitLogger(cfg)

		if logger.Enabled(context.Background(), slog.LevelInfo) {
			t.Error("Expected Info level to be disabled when Error level is set")
		}
		if !logger.Enabled(context.Background(), slog.LevelError) {
			t.Error("Expected Error level to be enabled")
		}
	})

	t.Run("Source information included at Error and Debug levels", func(t *testing.T) {
		// Define test cases
		testCases := []struct {
			logLevel     slog.Level
			messageLevel slog.Level
			expectSource bool
		}{
			{logLevel: slog.LevelDebug, messageLevel: slog.LevelDebug, expectSource: true},
			{logLevel: slog.LevelDebug, messageLevel: slog.LevelInfo, expectSource: false},
			{logLevel: slog.LevelDebug, messageLevel: slog.LevelWarn, expectSource: false},
			{logLevel: slog.LevelDebug, messageLevel: slog.LevelError, expectSource: true},
			{logLevel: slog.LevelInfo, messageLevel: slog.LevelDebug, expectSource: false}, // Not logged
			{logLevel: slog.LevelInfo, messageLevel: slog.LevelInfo, expectSource: false},
			{logLevel: slog.LevelInfo, messageLevel: slog.LevelError, expectSource: true},
		}

		for _, tc := range testCases {
			// Reset stdout
			w.Close()
			r, w, _ = os.Pipe()
			os.Stdout = w

			cfg := LogConfig{
				Level:      tc.logLevel,
				JSONOutput: true,
			}
			logger := InitLogger(cfg)

			logger.Log(context.Background(), tc.messageLevel, "test message")
			w.Close()

			var buf bytes.Buffer
			_, err := buf.ReadFrom(r)
			if err != nil {
				t.Fatal(err)
			}

			output := buf.String()

			// Check if the log was emitted based on level
			if output == "" {
				if tc.messageLevel < tc.logLevel {
					continue // Log was correctly not emitted
				}
				t.Errorf("Expected log output for level %v", tc.messageLevel)
				continue
			}

			var jsonMap map[string]interface{}
			if err := json.Unmarshal([]byte(output), &jsonMap); err != nil {
				t.Errorf("Expected valid JSON output, got error: %v", err)
				continue
			}

			_, hasSourceFile := jsonMap["source_file"]
			_, hasSourceLine := jsonMap["source_line"]
			_, hasFunction := jsonMap["function"]

			if tc.expectSource {
				if !hasSourceFile || !hasSourceLine || !hasFunction {
					t.Errorf("Expected source information in log for level %v", tc.messageLevel)
				}
			} else {
				if hasSourceFile || hasSourceLine || hasFunction {
					t.Errorf("Did not expect source information in log for level %v", tc.messageLevel)
				}
			}
		}
	})

	// Restore stdout
	os.Stdout = oldStdout
}
