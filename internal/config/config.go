package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

const (
	defaultAppEnv   = "development"
	defaultHTTPAddr = ":8080"
)

type Config struct {
	AppEnv      string
	DatabaseURL string
	RedisURL    string
	HTTPAddr    string
}

type target string

const (
	targetAPI    target = "api"
	targetWorker target = "worker"
)

func LoadAPI() (Config, error) {
	return load(targetAPI)
}

func LoadWorker() (Config, error) {
	return load(targetWorker)
}

func load(kind target) (Config, error) {
	cfg := Config{
		AppEnv:      stringValue("APP_ENV", defaultAppEnv),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisURL:    redisValue(),
	}

	if kind == targetAPI {
		cfg.HTTPAddr = strings.TrimSpace(os.Getenv("HTTP_ADDR"))
		if cfg.HTTPAddr == "" && cfg.AppEnv == defaultAppEnv {
			cfg.HTTPAddr = defaultHTTPAddr
		}
	}

	var validationErrors []string

	validationErrors = append(validationErrors, validateRequiredURL("DATABASE_URL", cfg.DatabaseURL)...)
	validationErrors = append(validationErrors, validateRequiredURL("REDIS_URL", cfg.RedisURL)...)

	if kind == targetAPI {
		switch {
		case cfg.HTTPAddr == "":
			validationErrors = append(validationErrors, "HTTP_ADDR is required")
		case !validTCPAddr(cfg.HTTPAddr):
			validationErrors = append(validationErrors, "HTTP_ADDR must be a valid TCP listen address")
		}
	}

	if len(validationErrors) > 0 {
		return Config{}, errors.New("invalid configuration:\n - " + strings.Join(validationErrors, "\n - "))
	}

	return cfg, nil
}

func stringValue(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func redisValue() string {
	if redisURL := strings.TrimSpace(os.Getenv("REDIS_URL")); redisURL != "" {
		return redisURL
	}

	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if redisAddr == "" {
		return ""
	}

	return "redis://" + redisAddr
}

func validateRequiredURL(name, value string) []string {
	if value == "" {
		return []string{fmt.Sprintf("%s is required", name)}
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return []string{fmt.Sprintf("%s must be a valid URL", name)}
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return []string{fmt.Sprintf("%s must be a valid URL", name)}
	}

	return nil
}

func validTCPAddr(value string) bool {
	_, err := net.ResolveTCPAddr("tcp", value)
	return err == nil
}
