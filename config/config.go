// Path: config/config.go

package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// BrokerConfig represents a single MQTT broker configuration
type BrokerConfig struct {
	Broker     string   `json:"broker"`
	Port       string   `json:"port"`
	Username   string   `json:"username,omitempty"`
	Password   string   `json:"password,omitempty"`
	Topics     []string `json:"topics"`
	OutputFile string   `json:"output_file,omitempty"` // Optional: separate file for this broker
}

// Config represents the application configuration
type Config struct {
	Brokers  []BrokerConfig `json:"brokers,omitempty"` // New: multiple brokers
	Detailed bool           `json:"detailed,omitempty"`

	// Legacy single-broker fields (for backward compatibility)
	Broker   string   `json:"-"`
	Port     string   `json:"-"`
	GateID   string   `json:"-"`
	Topics   []string `json:"-"`
	Username string   `json:"-"`
	Password string   `json:"-"`
	LogFile  string   `json:"-"`
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

	cfg := &Config{
		Detailed: getEnv("DETAILED", "false") == "true",
	}

	// Try to load from JSON config file first
	configFile := getEnv("CONFIG_FILE", "config.json")
	if data, err := os.ReadFile(configFile); err == nil {
		if err := json.Unmarshal(data, cfg); err == nil && len(cfg.Brokers) > 0 {
			// Successfully loaded from JSON, validate and return
			cfg = validateAndNormalizeConfig(cfg)
			return cfg
		}
	}

	// Fall back to environment variables (backward compatibility + new multi-broker support)
	cfg = loadConfigFromEnv()
	return cfg
}

// validateAndNormalizeConfig validates and normalizes broker configurations
func validateAndNormalizeConfig(cfg *Config) *Config {
	for i := range cfg.Brokers {
		broker := &cfg.Brokers[i]

		// Set defaults
		if broker.Port == "" {
			broker.Port = "1883"
		}

		// Default wildcard if no topics specified
		if len(broker.Topics) == 0 {
			broker.Topics = []string{"#"}
		}

		// Trim spaces from topics
		for j := range broker.Topics {
			broker.Topics[j] = strings.TrimSpace(broker.Topics[j])
		}
	}
	return cfg
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() *Config {
	cfg := &Config{
		Detailed: getEnv("DETAILED", "false") == "true",
	}

	// Check for multi-broker configuration
	// Format: BROKER_1_HOST, BROKER_1_PORT, BROKER_1_TOPICS, BROKER_1_OUTPUT_FILE, etc.
	brokerIndex := 1
	var brokers []BrokerConfig

	for {
		brokerHost := getEnv(fmt.Sprintf("BROKER_%d_HOST", brokerIndex), "")
		if brokerHost == "" {
			break // No more brokers
		}

		broker := BrokerConfig{
			Broker:     brokerHost,
			Port:       getEnv(fmt.Sprintf("BROKER_%d_PORT", brokerIndex), "1883"),
			Username:   getEnv(fmt.Sprintf("BROKER_%d_USERNAME", brokerIndex), ""),
			Password:   getEnv(fmt.Sprintf("BROKER_%d_PASSWORD", brokerIndex), ""),
			OutputFile: getEnv(fmt.Sprintf("BROKER_%d_OUTPUT_FILE", brokerIndex), ""),
		}

		// Load topics
		topicsEnv := getEnv(fmt.Sprintf("BROKER_%d_TOPICS", brokerIndex), "")
		if topicsEnv != "" {
			broker.Topics = strings.Split(topicsEnv, ",")
			for i := range broker.Topics {
				broker.Topics[i] = strings.TrimSpace(broker.Topics[i])
			}
		}

		// Check for gate_id for this broker
		gateID := getEnv(fmt.Sprintf("BROKER_%d_GATE_ID", brokerIndex), "")
		if gateID != "" {
			gateTopics := []string{
				fmt.Sprintf("%s_localfirst", gateID),
				fmt.Sprintf("%s_parkbox", gateID),
				fmt.Sprintf("%s_status", gateID),
				fmt.Sprintf("%s_events", gateID),
			}
			broker.Topics = append(broker.Topics, gateTopics...)
		}

		// Default wildcard if no topics specified
		if len(broker.Topics) == 0 {
			broker.Topics = []string{"#"}
		}

		brokers = append(brokers, broker)
		brokerIndex++
	}

	// If no multi-broker config found, fall back to legacy single-broker format
	if len(brokers) == 0 {
		broker := BrokerConfig{
			Broker:   getEnv("MQTT_BROKER", "localhost"),
			Port:     getEnv("MQTT_PORT", "1883"),
			Username: getEnv("MQTT_USERNAME", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
		}

		// Legacy log file handling
		logFileEnv := getEnv("LOG_FILE", "")
		switch logFileEnv {
		case "":
			broker.OutputFile = ""
		case "true", "1":
			broker.OutputFile = "mqtt.log"
		default:
			broker.OutputFile = logFileEnv
		}

		// Load topics from env
		topicsEnv := getEnv("MQTT_TOPICS", "")
		if topicsEnv != "" {
			broker.Topics = strings.Split(topicsEnv, ",")
			for i := range broker.Topics {
				broker.Topics[i] = strings.TrimSpace(broker.Topics[i])
			}
		}

		// If gate_id is provided, generate topics
		gateID := getEnv("GATE_ID", "")
		if gateID != "" {
			gateTopics := []string{
				fmt.Sprintf("%s_localfirst", gateID),
				fmt.Sprintf("%s_parkbox", gateID),
				fmt.Sprintf("%s_status", gateID),
				fmt.Sprintf("%s_events", gateID),
			}
			broker.Topics = append(broker.Topics, gateTopics...)
		}

		// Default wildcard if no topics specified
		if len(broker.Topics) == 0 {
			broker.Topics = []string{"#"}
		}

		brokers = append(brokers, broker)
	}

	cfg.Brokers = brokers
	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
