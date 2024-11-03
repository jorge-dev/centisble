package config

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"text/template"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jorge-dev/centsible/internal/version"
)

type Config struct {
	Port     int
	AppEnv   string
	Database DatabaseConfig
	JWT      JWTConfig
}

type DatabaseConfig struct {
	Host         string
	Port         string
	Database     string
	Username     string
	Password     string
	Schema       string
	RunMigration bool
}

type JWTConfig struct {
	Secret string
}

var (
	config *Config
	once   sync.Once
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
)

// Get returns the singleton instance of Config
func Get() *Config {
	once.Do(func() {
		config = &Config{
			Port:   loadPort(),
			AppEnv: loadEnvWithDefault("APP_ENV", "local"),
			Database: DatabaseConfig{
				Host:         requireEnv("CENTSIBLE_DB_HOST"),
				Port:         requireEnv("CENTSIBLE_DB_PORT"),
				Database:     requireEnv("CENTSIBLE_DB_DATABASE"),
				Username:     requireEnv("CENTSIBLE_DB_USERNAME"),
				Password:     requireEnv("CENTSIBLE_DB_PASSWORD"),
				Schema:       requireEnv("CENTSIBLE_DB_SCHEMA"),
				RunMigration: os.Getenv("RUN_MIGRATION") == "true",
			},
			JWT: JWTConfig{
				Secret: requireEnv("JWT_SECRET"),
			},
		}
	})
	return config
}

func loadPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 8080 // default port
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid PORT: %v", err)
	}
	return port
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}

func loadEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// BannerData holds information for the banner template
type BannerData struct {
	BannerColor        string
	TextColor          string
	Reset              string
	ApplicationTitle   string
	ApplicationVersion string
	CompileDate        string
	GitCommit          string
	GoVersion          string
	OS                 string
	Arch               string
	Version            string
}

// PrintBannerFromFile reads and prints an ASCII banner from a file with colors
func (c *Config) PrintBannerFromFile() {
	bannerTemplate, err := template.ParseFiles("./internal/config/banner.txt")
	if err != nil {
		log.Fatalf("Failed to parse banner template: %v", err)
	}

	versionInfo := version.Get()

	data := BannerData{
		BannerColor:        Blue,
		TextColor:          Green,
		Reset:              Reset,
		ApplicationTitle:   "Centsible",
		ApplicationVersion: versionInfo.GitVersion,
		CompileDate:        time.Now().UTC().Format(time.RFC1123),
		GitCommit:          versionInfo.GitCommit,
		GoVersion:          versionInfo.GoVersion,
		OS:                 runtime.GOOS,
		Arch:               runtime.GOARCH,
		Version:            versionInfo.GitVersion,
	}

	err = bannerTemplate.Execute(os.Stdout, data)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}
