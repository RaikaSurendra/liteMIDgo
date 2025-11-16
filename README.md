# LiteMIDgo

A lightweight middleware that acts as a web server proxy to ServiceNow instances, similar to ServiceNow's MID server but simpler and more focused on ECC queue operations.

## Features

- 🚀 **Lightweight**: Minimal resource footprint compared to full MID servers
- 🌐 **Web Server**: HTTP/HTTPS server with REST API endpoints
- 🔗 **ServiceNow Integration**: Direct communication with ServiceNow ECC Queue
- ⚙️ **Configurable**: YAML-based configuration with CLI setup
- 🛡️ **Secure**: Basic authentication and HTTPS support
- 📊 **Health Monitoring**: Built-in health checks and connection testing
- 🔄 **Graceful Shutdown**: Proper signal handling for production deployment

## Quick Start

### 1. Build the Application

```bash
go build -o litemidgo .
```

### 2. Configure ServiceNow Connection

```bash
./litemidgo config
```

This will guide you through setting up:
- ServiceNow instance URL
- Authentication credentials
- Server configuration
- Network settings

### 3. Test Connection

```bash
./litemidgo config test
```

### 4. Start the Server

```bash
./litemidgo server
```

The server will start on `http://localhost:8080` (or your configured host/port).

## API Endpoints

### Health Check
```bash
GET /health
```

Returns the health status of the service and ServiceNow connection.

### ECC Queue Proxy
```bash
POST /proxy/ecc_queue
Content-Type: application/json

{
  "agent": "litemidgo",
  "topic": "MIDServer", 
  "name": "default",
  "source": "client-ip",
  "payload": {
    "command": "your-command",
    "data": "your-data"
  }
}
```

### Server Information
```bash
GET /
```

Returns server information and available endpoints.

## Configuration

Configuration is managed through a YAML file (`config/config.yaml` by default):

```yaml
server:
  host: "0.0.0.0"
  port: 8080

servicenow:
  instance: "your-instance.service-now.com"
  username: "your-username"
  password: "your-password"
  use_https: true
  timeout: 30
```

### Configuration Locations

The application searches for configuration in this order:
1. Environment variables (take precedence)
2. `./config.yaml`
3. `./config/config.yaml`
4. `$HOME/.litemidgo/config.yaml`

### Environment Variables

You can configure ServiceNow credentials using environment variables. The application will automatically load them from a `.env` file if present, or you can set them directly in your shell.

**Option 1: Using .env file (recommended for development)**
```bash
# Copy the example file
cp .env.example .env

# Edit .env with your credentials
SERVICENOW_INSTANCE=your-instance.service-now.com
SERVICENOW_USERNAME=your-username
SERVICENOW_PASSWORD=your-password
```

**Option 2: Using shell environment variables (recommended for production)**
```bash
export SERVICENOW_INSTANCE="your-instance.service-now.com"
export SERVICENOW_USERNAME="your-username"
export SERVICENOW_PASSWORD="your-password"
```

Environment variables take precedence over config file settings and are more secure for production deployments. The `.env` file is automatically excluded from git via `.gitignore`.

## CLI Commands

### Server Management
```bash
# Start the server
./litemidgo server

# Start with custom config
./litemidgo server --config /path/to/config.yaml

# Enable debug logging
./litemidgo server --debug
```

### Configuration
```bash
# Interactive setup
./litemidgo config

# Test connection
./litemidgo config test
```

## Use Cases

### 1. Network Isolation
Deploy LiteMIDgo on a machine that has internet access (port 443) while other machines in your network can only reach the LiteMIDgo server.

### 2. ServiceNow Integration
Send data to ServiceNow ECC Queue from applications that don't have direct ServiceNow access.

### 3. Command Line Tools
Use curl or other CLI tools to send data to ServiceNow:

```bash
curl -X POST http://localhost:8080/proxy/ecc_queue \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "litemidgo",
    "topic": "MIDServer",
    "name": "test",
    "payload": {"message": "Hello from LiteMIDgo"}
  }'
```

## Architecture

```
[Client Applications] -> [LiteMIDgo Server] -> [ServiceNow Instance]
                                    |
                               ECC Queue
                                    |
                           [ServiceNow Sensors/Processors]
```

## Security Considerations

- Use HTTPS in production environments
- Secure the configuration file containing credentials
- Consider implementing additional authentication for the proxy endpoints
- Monitor access logs for unauthorized usage

## Development

### Project Structure
```
litemidgo/
├── main.go                 # Application entry point
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   ├── server.go          # Server command
│   └── config.go          # Configuration command
├── config/                # Configuration management
│   ├── config.go          # Config loading and validation
│   └── config.yaml        # Default configuration
├── internal/              # Internal packages
│   ├── server/            # HTTP server implementation
│   │   └── server.go
│   └── servicenow/        # ServiceNow client
│       └── client.go
└── go.mod                 # Go module definition
```

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

## License

This project is provided as-is for educational and development purposes.

## Contributing

Feel free to submit issues and enhancement requests!
