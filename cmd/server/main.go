// Path: cmd/main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mqtt/config"

	"mqtt/pkg/logger"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	if logger.IsDetailed() {
		payload := msg.Payload()
		// Try to pretty-print JSON if it's valid JSON
		var jsonObj interface{}
		var prettyPayload string
		if err := json.Unmarshal(payload, &jsonObj); err == nil {
			// It's valid JSON, pretty print it with 2-space indentation
			if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				prettyPayload = string(prettyJSON)
			} else {
				prettyPayload = string(payload)
			}
		} else {
			prettyPayload = string(payload)
		}
		// Output topic and formatted JSON
		logger.Info(fmt.Sprintf("📩 %s (%d bytes)\n%s", msg.Topic(), len(msg.Payload()), prettyPayload))
	} else {
		logger.Info(fmt.Sprintf("📩 %s (%d bytes)", msg.Topic(), len(msg.Payload())))
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	logger.Info("✓ Connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	logger.Error("✗ Connection lost", zap.Error(err))
}

func main() {
	cfg := config.LoadConfig()

	// Initialize logger with optional file logging
	logger.InitLoggerWithFile(cfg.LogFile)
	defer func() {
		// Only sync on shutdown, not on every log
		_ = logger.Sync()
	}()

	logger.SetDetailed(cfg.Detailed)

	logger.Info("🚀 MQTT Subscriber Service")
	logger.Info(fmt.Sprintf("📡 Broker: %s:%s", cfg.Broker, cfg.Port))
	logger.Info(fmt.Sprintf("📋 Topics: %s", strings.Join(cfg.Topics, ", ")))
	logger.Info(fmt.Sprintf("🔍 Detailed Mode: %v", cfg.Detailed))
	if cfg.LogFile != "" {
		logger.Info(fmt.Sprintf("📝 Logging to file: %s", cfg.LogFile))
	}
	logger.Info(strings.Repeat("─", 80))

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", cfg.Broker, cfg.Port))
	opts.SetClientID(fmt.Sprintf("mqtt_sub_%d", time.Now().Unix()))
	opts.SetDefaultPublishHandler(messageHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logger.Fatal("❌ Failed to connect", zap.Error(token.Error()))
	}

	// Subscribe to all topics
	for _, topic := range cfg.Topics {
		if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
			logger.Error("❌ Failed to subscribe", zap.String("topic", topic), zap.Error(token.Error()))
		} else {
			logger.Info(fmt.Sprintf("✓ Subscribed: %s", topic))
		}
	}

	logger.Info(strings.Repeat("─", 80))
	logger.Info("👂 Listening for messages... (Press Ctrl+C to exit)")
	logger.Info(strings.Repeat("─", 80))

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("👋 Shutting down...")
	client.Disconnect(250)
}
