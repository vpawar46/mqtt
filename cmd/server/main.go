// Path: cmd/main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"mqtt/config"
	"mqtt/pkg/logger"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// BrokerConnection manages a single MQTT broker connection
type BrokerConnection struct {
	Config     config.BrokerConfig
	Client     mqtt.Client
	OutputFile *os.File
	mu         sync.Mutex
}

// createMessageHandler creates a message handler for a specific broker
func createMessageHandler(broker *BrokerConnection, brokerIndex int) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		broker.mu.Lock()
		defer broker.mu.Unlock()

		brokerAddr := fmt.Sprintf("%s:%s", broker.Config.Broker, broker.Config.Port)
		message := formatMessage(msg, logger.IsDetailed())

		// Log to stdout with broker identifier
		logger.Info(fmt.Sprintf("[Broker %d: %s] %s", brokerIndex+1, brokerAddr, message))

		// Write to broker-specific file if configured
		if broker.OutputFile != nil {
			timestamp := time.Now().Format("2006-01-02 15:04:05.000")
			fileMessage := fmt.Sprintf("[%s] [Broker %d: %s] %s\n", timestamp, brokerIndex+1, brokerAddr, message)
			broker.OutputFile.WriteString(fileMessage)
			broker.OutputFile.Sync()
		}
	}
}

// formatMessage formats the MQTT message for output
func formatMessage(msg mqtt.Message, detailed bool) string {
	if detailed {
		payload := msg.Payload()
		var jsonObj interface{}
		var prettyPayload string
		if err := json.Unmarshal(payload, &jsonObj); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				prettyPayload = string(prettyJSON)
			} else {
				prettyPayload = string(payload)
			}
		} else {
			prettyPayload = string(payload)
		}
		return fmt.Sprintf("📩 %s (%d bytes)\n%s", msg.Topic(), len(msg.Payload()), prettyPayload)
	}
	return fmt.Sprintf("📩 %s (%d bytes)", msg.Topic(), len(msg.Payload()))
}

func main() {
	cfg := config.LoadConfig()

	// Initialize logger (stdout only, broker-specific files handled separately)
	logger.InitLogger()
	defer func() {
		_ = logger.Sync()
	}()

	logger.SetDetailed(cfg.Detailed)

	logger.Info("🚀 MQTT Subscriber Service")
	logger.Info(fmt.Sprintf("📊 Configured Brokers: %d", len(cfg.Brokers)))
	logger.Info(fmt.Sprintf("🔍 Detailed Mode: %v", cfg.Detailed))
	logger.Info(strings.Repeat("─", 80))

	if len(cfg.Brokers) == 0 {
		logger.Fatal("❌ No brokers configured")
	}

	// Initialize all broker connections
	connections := make([]*BrokerConnection, 0, len(cfg.Brokers))

	for i, brokerCfg := range cfg.Brokers {
		broker := &BrokerConnection{
			Config: brokerCfg,
		}

		// Open output file if specified
		if brokerCfg.OutputFile != "" {
			// If path doesn't contain directory separator, put it in logs/ directory
			outputPath := brokerCfg.OutputFile
			if !filepath.IsAbs(outputPath) && filepath.Dir(outputPath) == "." {
				outputPath = filepath.Join("logs", outputPath)
			}

			// Create logs directory if it doesn't exist
			dir := filepath.Dir(outputPath)
			if dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					logger.Error("❌ Failed to create log directory", zap.String("dir", dir), zap.Error(err))
				}
			}

			file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				logger.Error("❌ Failed to open output file", zap.String("file", outputPath), zap.Error(err))
			} else {
				broker.OutputFile = file
				logger.Info(fmt.Sprintf("📝 Broker %d: Logging to file: %s", i+1, outputPath))
			}
		}

		// Create MQTT client options
		brokerAddr := fmt.Sprintf("%s:%s", brokerCfg.Broker, brokerCfg.Port)
		opts := mqtt.NewClientOptions()
		opts.AddBroker(fmt.Sprintf("tcp://%s", brokerAddr))
		opts.SetClientID(fmt.Sprintf("mqtt_sub_%d_%d", time.Now().Unix(), i))
		opts.SetDefaultPublishHandler(createMessageHandler(broker, i))
		opts.OnConnect = createConnectHandler(i, brokerAddr)
		opts.OnConnectionLost = createConnectionLostHandler(i, brokerAddr)
		opts.SetAutoReconnect(true)
		opts.SetCleanSession(true)

		if brokerCfg.Username != "" {
			opts.SetUsername(brokerCfg.Username)
			opts.SetPassword(brokerCfg.Password)
		}

		broker.Client = mqtt.NewClient(opts)

		// Connect to broker
		logger.Info(fmt.Sprintf("🔌 Connecting to Broker %d: %s...", i+1, brokerAddr))
		if token := broker.Client.Connect(); token.Wait() && token.Error() != nil {
			logger.Error("❌ Failed to connect", zap.Int("broker", i+1), zap.String("address", brokerAddr), zap.Error(token.Error()))
			if broker.OutputFile != nil {
				broker.OutputFile.Close()
			}
			continue
		}

		// Subscribe to all topics for this broker
		for _, topic := range brokerCfg.Topics {
			if token := broker.Client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
				logger.Error("❌ Failed to subscribe", zap.Int("broker", i+1), zap.String("topic", topic), zap.Error(token.Error()))
			} else {
				logger.Info(fmt.Sprintf("✓ Broker %d: Subscribed to %s", i+1, topic))
			}
		}

		connections = append(connections, broker)
	}

	if len(connections) == 0 {
		logger.Fatal("❌ No successful broker connections")
	}

	logger.Info(strings.Repeat("─", 80))
	logger.Info(fmt.Sprintf("👂 Listening for messages from %d broker(s)... (Press Ctrl+C to exit)", len(connections)))
	logger.Info(strings.Repeat("─", 80))

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("👋 Shutting down...")

	// Disconnect all clients and close all files
	for i, broker := range connections {
		if broker.Client != nil && broker.Client.IsConnected() {
			broker.Client.Disconnect(250)
			logger.Info(fmt.Sprintf("✓ Broker %d disconnected", i+1))
		}
		if broker.OutputFile != nil {
			broker.OutputFile.Close()
		}
	}
}

// createConnectHandler creates a connect handler for a specific broker
func createConnectHandler(brokerIndex int, brokerAddr string) mqtt.OnConnectHandler {
	return func(client mqtt.Client) {
		logger.Info(fmt.Sprintf("✓ Broker %d: Connected to %s", brokerIndex+1, brokerAddr))
	}
}

// createConnectionLostHandler creates a connection lost handler for a specific broker
func createConnectionLostHandler(brokerIndex int, brokerAddr string) mqtt.ConnectionLostHandler {
	return func(client mqtt.Client, err error) {
		logger.Error(fmt.Sprintf("✗ Broker %d: Connection lost to %s", brokerIndex+1, brokerAddr), zap.Error(err))
	}
}
