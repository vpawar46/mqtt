# How to View Log Files

## Important: Understanding Files

- **`mqtt-subscriber`** = Binary executable (compiled program) - NOT a log file
- **Log files** = Text files created when you run with `LOG_FILE` environment variable

## Creating a Log File

Run the program with `LOG_FILE` environment variable. Log files are automatically saved in the `logs/` directory:

```bash
# Create a log file named mqtt.log in logs/ directory
LOG_FILE=mqtt.log ./mqtt-subscriber

# Or with detailed mode
LOG_FILE=mqtt.log DETAILED=true ./mqtt-subscriber

# Or specify full path
LOG_FILE=logs/custom.log ./mqtt-subscriber
```

**Note:** The `logs/` directory is automatically created if it doesn't exist.

## Viewing Log Files

### Method 1: Using `cat` (view entire file)
```bash
cat logs/mqtt.log
```

### Method 2: Using `less` (scrollable view)
```bash
less logs/mqtt.log
# Press 'q' to quit, arrow keys to scroll
```

### Method 3: Using `tail` (watch live updates)
```bash
# View last 50 lines
tail -n 50 logs/mqtt.log

# Follow live updates (like tail -f)
tail -f logs/mqtt.log
```

### Method 4: Using `head` (view first lines)
```bash
head -n 50 logs/mqtt.log
```

### Method 5: Using `grep` (search for specific content)
```bash
# Search for specific topic
grep "2305_localfirst" logs/mqtt.log

# Search with context (5 lines before/after)
grep -C 5 "2305_localfirst" logs/mqtt.log
```

### Method 6: Using text editors
```bash
# Nano (simple editor)
nano logs/mqtt.log

# Vim
vim logs/mqtt.log

# Emacs
emacs logs/mqtt.log
```

## Log File Format

Log files are plain text with the following format:
```
2025-01-07 15:37:57.345	info	🚀 MQTT Subscriber Service
2025-01-07 15:37:57.345	info	📡 Broker: 172.27.59.101:8883
2025-01-07 15:37:57.476	info	✓ Connected to MQTT broker
2025-01-07 15:37:57.648	info	📩 2305_localfirst (258 bytes)
{
  "sender_mac_add": "40:22:d8:1e:c0:27",
  "receiver_mac": "ff:ff:ff:ff:ff:ff",
  ...
}
```

## Quick Examples

```bash
# 1. Run with logging (creates logs/mqtt.log)
LOG_FILE=mqtt.log ./mqtt-subscriber

# 2. In another terminal, watch the log file
tail -f logs/mqtt.log

# 3. Search for errors
grep -i error logs/mqtt.log

# 4. Count messages
grep "📩" logs/mqtt.log | wc -l

# 5. List all log files
ls -lh logs/
```

