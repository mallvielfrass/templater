package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	devJWTSecret   = "dev-jwt-secret-change-me"
	devOOJWTSecret = "onlyoffice-dev-secret"
)

type Config struct {
	HttpPort            int
	BadgerDBPath        string
	JWTSecret           string
	PublicBaseURL       string
	OnlyOfficeJWTSecret string
	OnlyOfficeURL       string
	CORSOrigins         []string
	StaticDir           string
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func envOr(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func isProduction() bool {
	e := strings.ToLower(os.Getenv("GO_ENV"))
	return e == "production" || e == "prod"
}

func NewConfig(configPath string) (*Config, error) {
	var conf Config
	if fileExists(configPath) {
		err := godotenv.Load(configPath)
		if err != nil {
			return nil, err
		}
	}

	httpPort := 3053
	if v := os.Getenv("HttpPort"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("HttpPort: %w", err)
		}
		httpPort = parsed
	}
	conf.HttpPort = httpPort
	conf.BadgerDBPath = envOr("BadgerDBPath", "./data/badger")
	conf.JWTSecret = envOr("JWTSecret", devJWTSecret)
	conf.PublicBaseURL = envOr("PublicBaseURL", "http://host.docker.internal:3053")
	conf.OnlyOfficeJWTSecret = envOr("OnlyOfficeJWTSecret", devOOJWTSecret)
	conf.OnlyOfficeURL = envOr("OnlyOfficeURL", "http://127.0.0.1:8080")
	conf.CORSOrigins = splitCSV(envOr("CORSOrigins", "http://localhost:5173"))
	conf.StaticDir = envOr("StaticDir", "frontend/dist")

	return &conf, nil
}

func (c *Config) Validate() error {
	if c.HttpPort == 0 {
		return fmt.Errorf("http port is not set")
	}
	if c.BadgerDBPath == "" {
		return fmt.Errorf("badger db path is not set")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("jwt secret is not set")
	}
	if c.PublicBaseURL == "" {
		return fmt.Errorf("public base url is not set")
	}
	if c.OnlyOfficeJWTSecret == "" {
		return fmt.Errorf("onlyoffice jwt secret is not set")
	}
	if isProduction() {
		if c.JWTSecret == devJWTSecret {
			return fmt.Errorf("jwt secret must not use the dev default in production")
		}
		if c.OnlyOfficeJWTSecret == devOOJWTSecret {
			return fmt.Errorf("onlyoffice jwt secret must not use the dev default in production")
		}
	}
	return nil
}
