# MQTT Subscriber Service

A lightweight, efficient MQTT subscriber for troubleshooting and monitoring MQTT messages. Replaces MQTT Explorer for quick debugging.

## Features

- ✅ Subscribe to multiple topics
- ✅ Auto-generate topics from gate_id
- ✅ Clean, timestamped message logging
- ✅ Optional file logging for persistence
- ✅ Auto-reconnect on connection loss
- ✅ Environment-based configuration
- ✅ Wildcard topic support

## Installation

```bash
go mod download
go build -o mqtt-sub
```

## Usage

### Quick Start

```bash
# Subscribe to all topics on localhost
./mqtt-sub

# With environment variables
MQTT_BROKER=192.168.1.100 MQTT_PORT=1883 ./mqtt-sub
```

### Configuration Options

**Option 1: Use Gate ID**
```bash
GATE_ID=23 ./mqtt-sub
# Subscribes to: 23_localfirst, 23_parkbox, 23_status, 23_events
```

**Option 2: Specify Topics**
```bash
MQTT_TOPICS="sensor/#,alerts/#,status/+" ./mqtt-sub
```

**Option 3: Use .env file**
```bash
cp .env.example .env
# Edit .env with your settings
./mqtt-sub
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MQTT_BROKER` | localhost | MQTT broker hostname/IP |
| `MQTT_PORT` | 1883 | MQTT broker port |
| `GATE_ID` | - | Gate ID for auto-topic generation |
| `MQTT_TOPICS` | # | Comma-separated topic list |
| `MQTT_USERNAME` | - | MQTT username (if required) |
| `MQTT_PASSWORD` | - | MQTT password (if required) |
| `DETAILED` | false | Enable detailed JSON payload logging |
| `LOG_FILE` | - | Log file name (saved in `logs/` directory) or full path |

## Examples

**Monitor specific gate:**
```bash
MQTT_BROKER=broker.example.com GATE_ID=23 ./mqtt-sub
```

**Custom topics:**
```bash
MQTT_TOPICS="home/sensors/#,alerts/critical" ./mqtt-sub
```

**With authentication:**
```bash
MQTT_BROKER=secure.broker.com \
MQTT_USERNAME=admin \
MQTT_PASSWORD=secret \
MQTT_TOPICS="devices/#" \
./mqtt-sub
```

**With file logging:**
```bash
# Log to both stdout and file (saved in logs/ directory)
LOG_FILE=mqtt.log DETAILED=true ./mqtt-sub

# Or use full path
LOG_FILE=logs/mqtt.log DETAILED=true ./mqtt-sub

# Or use .env file
echo "LOG_FILE=mqtt.log" >> .env
echo "DETAILED=true" >> .env
./mqtt-sub
```

**Viewing log files:**
```bash
# View entire file (logs are in logs/ directory)
cat logs/mqtt.log

# Watch live updates
tail -f logs/mqtt.log

# View last 50 lines
tail -n 50 logs/mqtt.log

# Search for specific topic
grep "2305_localfirst" logs/mqtt.log
```
See `LOG_VIEWING.md` for complete viewing instructions.

## Output Format

```
[15:04:05.000] 23_localfirst
{"event": "gate_open", "timestamp": 1234567890}
--------------------------------------------------------------------------------
[15:04:06.123] 23_parkbox
{"status": "occupied", "vehicle_id": "ABC123"}
--------------------------------------------------------------------------------
```

## Building for Production

```bash
# Linux
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mqtt-sub

# Optimized build
go build -ldflags="-s -w" -o mqtt-sub
```

## Running as Service (systemd)

Create `/etc/systemd/system/mqtt-sub.service`:

```ini
[Unit]
Description=MQTT Subscriber Service
After=network.target

[Service]
Type=simple
User=mqtt
WorkingDirectory=/opt/mqtt-sub
EnvironmentFile=/opt/mqtt-sub/.env
ExecStart=/opt/mqtt-sub/mqtt-sub
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable mqtt-sub
sudo systemctl start mqtt-sub
sudo systemctl status mqtt-sub
```

## Performance

- **Memory**: ~10MB base
- **CPU**: Minimal (event-driven)
- **Concurrent**: Handles thousands of messages/sec
- **Reconnect**: Automatic with exponential backoff

## License

MIT
