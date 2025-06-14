package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	LogLevel    string
	DockerHost  string
	MetricsEnabled bool
}

func New() *Config {
	return &Config{
		Port:           getEnvInt("LOCALCLOUD_PORT", 8080),
		LogLevel:       getEnv("LOCALCLOUD_LOG_LEVEL", "INFO"),
		DockerHost:     getEnv("DOCKER_HOST", ""),
		MetricsEnabled: getEnvBool("LOCALCLOUD_METRICS", true),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
