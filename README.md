# MQTT Subscriber Service

A lightweight, efficient MQTT subscriber for troubleshooting and monitoring MQTT messages. Replaces MQTT Explorer for quick debugging.

## Features

- ✅ **Multiple brokers support** - Connect to multiple MQTT brokers simultaneously
- ✅ **Per-broker topics** - Each broker can subscribe to different topics
- ✅ **Separate output files** - Each broker can write to its own log file
- ✅ Subscribe to multiple topics per broker
- ✅ Auto-generate topics from gate_id
- ✅ Clean, timestamped message logging
- ✅ Optional file logging for persistence
- ✅ Auto-reconnect on connection loss
- ✅ Environment-based or JSON configuration
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

**Option 1: JSON Configuration File (Recommended for Multiple Brokers)**
```bash
# Create config.json from example
cp config.json.example config.json
# Edit config.json with your broker settings
./mqtt-sub

# Or specify custom config file
CONFIG_FILE=my-config.json ./mqtt-sub
```

**Option 2: Environment Variables (Multiple Brokers)**
```bash
# Broker 1
BROKER_1_HOST=172.27.59.101
BROKER_1_PORT=8883
BROKER_1_USERNAME=user1
BROKER_1_PASSWORD=pass1
BROKER_1_TOPICS="2309_tag_status,2309_dashboard"
BROKER_1_OUTPUT_FILE=broker1.log

# Broker 2
BROKER_2_HOST=172.27.66.110
BROKER_2_PORT=8883
BROKER_2_TOPICS="2428_dashboard"
BROKER_2_OUTPUT_FILE=broker2.log

./mqtt-sub
```

**Option 3: Legacy Single Broker (Backward Compatible)**
```bash
# Use Gate ID
GATE_ID=23 ./mqtt-sub
# Subscribes to: 23_localfirst, 23_parkbox, 23_status, 23_events

# Or specify topics
MQTT_TOPICS="sensor/#,alerts/#,status/+" ./mqtt-sub
```

**Option 4: Use .env file**
```bash
cp .env.example .env
# Edit .env with your settings
./mqtt-sub
```

### Configuration Variables

#### JSON Configuration File
Create a `config.json` file (see `config.json.example`):
```json
{
  "detailed": true,
  "brokers": [
    {
      "broker": "172.27.59.101",
      "port": "8883",
      "username": "user1",
      "password": "pass1",
      "topics": ["2309_tag_status", "2309_dashboard"],
      "output_file": "broker1.log"
    }
  ]
}
```

#### Environment Variables

**Multi-Broker Configuration:**
| Variable | Description |
|----------|-------------|
| `BROKER_N_HOST` | Broker N hostname/IP (N = 1, 2, 3, ...) |
| `BROKER_N_PORT` | Broker N port (default: 1883) |
| `BROKER_N_USERNAME` | Broker N username (optional) |
| `BROKER_N_PASSWORD` | Broker N password (optional) |
| `BROKER_N_TOPICS` | Comma-separated topics for broker N |
| `BROKER_N_GATE_ID` | Gate ID for auto-topic generation for broker N |
| `BROKER_N_OUTPUT_FILE` | Output file for broker N (saved in `logs/` directory) |
| `CONFIG_FILE` | Path to JSON config file (default: `config.json`) |
| `DETAILED` | Enable detailed JSON payload logging (applies to all brokers) |

**Legacy Single-Broker Configuration (Backward Compatible):**
| Variable | Default | Description |
|----------|---------|-------------|
| `MQTT_BROKER` | localhost | MQTT broker hostname/IP |
| `MQTT_PORT` | 1883 | MQTT broker port |
| `GATE_ID` | - | Gate ID for auto-topic generation |
| `MQTT_TOPICS` | # | Comma-separated topic list |
| `MQTT_USERNAME` | - | MQTT username (if required) |
| `MQTT_PASSWORD` | - | MQTT password (if required) |
| `LOG_FILE` | - | Log file name (saved in `logs/` directory) or full path |

## Examples

### Multiple Brokers

**Using JSON config:**
```bash
# Create config.json with multiple brokers
cp config.json.example config.json
# Edit config.json
./mqtt-sub
```

**Using environment variables:**
```bash
BROKER_1_HOST=172.27.59.101 \
BROKER_1_PORT=8883 \
BROKER_1_TOPICS="2309_tag_status,2309_dashboard" \
BROKER_1_OUTPUT_FILE=broker1.log \
BROKER_2_HOST=172.27.66.110 \
BROKER_2_PORT=8883 \
BROKER_2_TOPICS="2428_dashboard" \
BROKER_2_OUTPUT_FILE=broker2.log \
DETAILED=true \
./mqtt-sub
```

### Single Broker (Legacy)

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
```

**Viewing log files:**
```bash
# View entire file (logs are in logs/ directory)
cat logs/mqtt.log

# Watch live updates (single broker)
tail -f logs/mqtt.log

# Watch multiple broker logs simultaneously
tail -f logs/broker1.log logs/broker2.log

# View last 50 lines
tail -n 50 logs/mqtt.log

# Search for specific topic across all logs
grep "2305_localfirst" logs/*.log
```
See `LOG_VIEWING.md` for complete viewing instructions.

## Output Format

**Console Output (with broker identifier):**
```
[16:47:07.636] [Broker 1: 172.27.59.101:8883] 📩 2309_tag_status (172 bytes)
{"message_body": {...}, "message_type_code": "6021"}
[16:47:08.123] [Broker 2: 172.27.66.110:8883] 📩 2428_dashboard (256 bytes)
{"status": "active", "data": {...}}
```

**File Output (per-broker files):**
Each broker writes to its configured output file with timestamps:
```
[2026-02-10 16:47:07.636] [Broker 1: 172.27.59.101:8883] 📩 2309_tag_status (172 bytes)
{"message_body": {...}}
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
