// Path: config/config.go

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Broker   string
	Port     string
	GateID   string
	Topics   []string
	Username string
	Password string
	Detailed bool
	LogFile  string
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // .env file is optional
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Only set if not already in environment
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

func LoadConfig() *Config {
	// Load .env file first
	loadEnvFile(".env")

	// Default log file to logs/mqtt.log if LOG_FILE is set but empty, or use provided value
	logFileEnv := getEnv("LOG_FILE", "")
	switch logFileEnv {
	case "":
		logFileEnv = "" // No logging by default
	case "true", "1":
		// If LOG_FILE is just "true" or "1", use default path
		logFileEnv = "mqtt.log"
	}

	cfg := &Config{
		Broker:   getEnv("MQTT_BROKER", "localhost"),
		Port:     getEnv("MQTT_PORT", "1883"),
		GateID:   getEnv("GATE_ID", ""),
		Username: getEnv("MQTT_USERNAME", ""),
		Password: getEnv("MQTT_PASSWORD", ""),
		Detailed: getEnv("DETAILED", "false") == "true",
		LogFile:  logFileEnv,
	}

	// Load topics from env
	topicsEnv := getEnv("MQTT_TOPICS", "")
	if topicsEnv != "" {
		cfg.Topics = strings.Split(topicsEnv, ",")
		for i := range cfg.Topics {
			cfg.Topics[i] = strings.TrimSpace(cfg.Topics[i])
		}
	}

	// If gate_id is provided, generate topics
	if cfg.GateID != "" {
		gateTopics := []string{
			fmt.Sprintf("%s_localfirst", cfg.GateID),
			fmt.Sprintf("%s_parkbox", cfg.GateID),
			fmt.Sprintf("%s_status", cfg.GateID),
			fmt.Sprintf("%s_events", cfg.GateID),
		}
		cfg.Topics = append(cfg.Topics, gateTopics...)
	}

	// Default wildcard if no topics specified
	if len(cfg.Topics) == 0 {
		cfg.Topics = []string{"#"}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
