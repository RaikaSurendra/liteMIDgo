#!/bin/bash

# LiteMIDgo API Examples
# Make sure the server is running before executing these commands

BASE_URL="http://localhost:8080"

echo "LiteMIDgo API Examples"
echo "======================="
echo

# Health Check
echo "1. Health Check:"
curl -s "$BASE_URL/health" | jq '.'
echo
echo

# Server Information
echo "2. Server Information:"
curl -s "$BASE_URL/" | jq '.'
echo
echo

# Send to ECC Queue - Simple Message
echo "3. Send Simple Message to ECC Queue:"
curl -s -X POST "$BASE_URL/proxy/ecc_queue" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "litemidgo",
    "topic": "MIDServer",
    "name": "test",
    "source": "curl-example",
    "payload": {
      "message": "Hello from LiteMIDgo",
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    }
  }' | jq '.'
echo
echo

# Send to ECC Queue - Command Execution
echo "4. Send Command to ECC Queue:"
curl -s -X POST "$BASE_URL/proxy/ecc_queue" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "litemidgo",
    "topic": "MIDServer",
    "name": "command",
    "source": "curl-example",
    "payload": {
      "command": "echo",
      "parameters": ["Hello", "World"],
      "working_directory": "/tmp"
    }
  }' | jq '.'
echo
echo

# Send to ECC Queue - File Operations
echo "5. Send File Operation to ECC Queue:"
curl -s -X POST "$BASE_URL/proxy/ecc_queue" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "litemidgo",
    "topic": "MIDServer",
    "name": "file_operation",
    "source": "curl-example",
    "payload": {
      "operation": "read",
      "file_path": "/etc/hosts",
      "encoding": "base64"
    }
  }' | jq '.'
echo
echo

echo "Examples completed!"
echo "Note: Make sure jq is installed for pretty JSON output"
