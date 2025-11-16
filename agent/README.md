# LiteMIDgo Agent

A SensuGo-compatible monitoring agent that collects system metrics and sends them to LiteMIDgo for ServiceNow integration.

## Features

- 🖥️ **System Metrics**: CPU, memory, disk, network, and runtime information
- 🌐 **Network Monitoring**: Interface statistics and active connections
- 📊 **SensuGo Compatible**: JSON output format compatible with SensuGo expectations
- 🔄 **Daemon Mode**: Continuous monitoring with configurable intervals
- 📡 **LiteMIDgo Integration**: Sends data to ServiceNow via LiteMIDgo ECC Queue

## Quick Start

### 1. Build the Agent

```bash
cd agent
go build -o litemidgo-agent .
```

### 2. Collect Metrics (Display Only)

```bash
./litemidgo-agent collect
```

### 3. Send Metrics Once

```bash
./litemidgo-agent send --server http://localhost:8080
```

### 4. Run as Daemon

```bash
./litemidgo-agent daemon --server http://localhost:8080 --interval 60
```

## Commands

### `collect`
Collect and display system metrics in JSON format.

```bash
./litemidgo-agent collect
```

**Output:**
```json
{
  "timestamp": "2025-11-16T19:15:00Z",
  "hostname": "macbook-pro",
  "os": {
    "platform": "darwin",
    "platform_family": "Standalone Workstation",
    "platform_version": "14.2.1",
    "architecture": "arm64",
    "kernel_version": "23.2.0",
    "virtualization_role": "guest"
  },
  "cpu": {
    "model_name": "Apple M3 Pro",
    "cores": 12,
    "logical_cores": 12,
    "usage_percent": 15.5,
    "load_average": [0.1, 0.2, 0.3],
    "frequency_mhz": 1.0
  },
  "memory": {
    "total": 34359738368,
    "available": 21474836480,
    "used": 12884901888,
    "used_percent": 37.5,
    "swap_total": 2147483648,
    "swap_used": 0,
    "swap_percent": 0.0
  },
  "disk": [
    {
      "device": "/dev/disk1s1",
      "mountpoint": "/",
      "fstype": "apfs",
      "total": 999320936448,
      "free": 534773774336,
      "used": 464547162112,
      "used_percent": 46.5
    }
  ],
  "network": {
    "interfaces": [
      {
        "name": "en0",
        "hardware_addr": "aa:bb:cc:dd:ee:ff",
        "mtu": 1500,
        "flags": ["up|broadcast|multicast"],
        "addresses": ["192.168.1.100/24"],
        "bytes_sent": 1234567890,
        "bytes_recv": 9876543210,
        "packets_sent": 1234567,
        "packets_recv": 8765432,
        "errin": 0,
        "errout": 0,
        "dropin": 0,
        "dropout": 0
      }
    ],
    "connections": [
      {
        "local_addr": "192.168.1.100:8080",
        "remote_addr": "192.168.1.1:53",
        "state": "ESTABLISHED",
        "pid": 1234,
        "process": "process_1234"
      }
    ]
  },
  "runtime": {
    "go_version": "go1.21.0",
    "go_os": "darwin",
    "go_arch": "arm64",
    "num_goroutine": 5,
    "num_cpu": 12
  }
}
```

### `send`
Collect and send metrics to LiteMIDgo server.

```bash
./litemidgo-agent send --server http://localhost:8080
```

**Payload sent to LiteMIDgo:**
```json
{
  "agent": "litemidgo-agent",
  "topic": "endpointData",
  "name": "macbook-pro",
  "source": "macbook-pro",
  "payload": {
    "metrics": { ... },
    "agent_version": "1.0.0",
    "collection_time": "2025-11-16T19:15:00Z"
  }
}
```

### `daemon`
Run continuously, sending metrics at regular intervals.

```bash
./litemidgo-agent daemon --server http://localhost:8080 --interval 60
```

## Configuration

### Command Line Options

- `--server, -s`: LiteMIDgo server URL (default: `http://localhost:8080`)
- `--interval, -i`: Collection interval in seconds (default: `60`)
- `--once`: Send metrics once and exit (daemon mode only)

### Environment Variables

```bash
export LITEMIDGO_SERVER="http://localhost:8080"
export LITEMIDGO_INTERVAL="30"
```

## Integration with ServiceNow

The agent sends metrics to LiteMIDgo with the topic `endpointData`, which gets queued in ServiceNow's ECC Queue. ServiceNow sensors can then process this data for:

- 📈 Performance monitoring
- 🚨 Alert generation
- 📊 Historical analysis
- 🔧 Automated remediation

## Example ServiceNow Sensor

```javascript
// ServiceNow ECC Queue Sensor Example
(function processECCQueue() {
    var eccQueueGR = new GlideRecord('ecc_queue');
    eccQueueGR.addQuery('topic', 'endpointData');
    eccQueueGR.addQuery('state', 'ready');
    eccQueueGR.orderByDesc('sys_created_on');
    eccQueueGR.query();
    
    while (eccQueueGR.next()) {
        try {
            var payload = JSON.parse(eccQueueGR.payload.toString());
            var metrics = payload.metrics;
            
            // Process metrics
            gs.info('Processing metrics from: ' + eccQueueGR.source);
            
            // Example: Create performance record
            var perfGR = new GlideRecord('cmdb_perf_metric');
            perfGr.initialize();
            perfGr.name = metrics.hostname + '_system_metrics';
            perfGr.value = JSON.stringify(metrics);
            perfGr.insert();
            
            // Mark as processed
            eccQueueGR.state = 'processed';
            eccQueueGR.update();
            
        } catch (ex) {
            gs.error('Error processing ECC Queue record: ' + ex.message);
            eccQueueGR.error_string = ex.message;
            eccQueueGR.state = 'error';
            eccQueueGR.update();
        }
    }
})();
```

## Deployment

### Systemd Service

```ini
# /etc/systemd/system/litemidgo-agent.service
[Unit]
Description=LiteMIDgo Agent
After=network.target

[Service]
Type=simple
User=litemidgo
WorkingDirectory=/opt/litemidgo-agent
ExecStart=/opt/litemidgo-agent/litemidgo-agent daemon --server http://litemidgo:8080 --interval 60
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable litemidgo-agent
sudo systemctl start litemidgo-agent
sudo systemctl status litemidgo-agent
```

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o litemidgo-agent .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/litemidgo-agent .
CMD ["./litemidgo-agent", "daemon", "--server", "http://litemidgo:8080", "--interval", "60"]
```

## Security Considerations

- 🔒 Use HTTPS in production environments
- 🔑 Configure proper authentication between agent and LiteMIDgo
- 🛡️ Monitor agent logs for unauthorized access attempts
- 📝 Limit network access to only required endpoints

## Troubleshooting

### Common Issues

1. **Connection refused**: Ensure LiteMIDgo server is running
2. **Permission denied**: Check network connectivity and firewall settings
3. **JSON parsing errors**: Verify LiteMIDgo server is running the correct version

### Debug Mode

```bash
# Enable verbose logging
LITEMIDGO_DEBUG=true ./litemidgo-agent daemon --server http://localhost:8080
```

## License

This project is provided as-is for educational and development purposes.
