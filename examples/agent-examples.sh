#!/bin/bash

# LiteMIDgo API Examples with Agent Integration
# Make sure the server is running before executing these commands

BASE_URL="http://localhost:8080"
AGENT_PATH="../agent/litemidgo-agent"

echo "LiteMIDgo API Examples with Agent Integration"
echo "=============================================="
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

# Agent - Collect Metrics Only
echo "3. Agent - Collect System Metrics:"
if [ -f "$AGENT_PATH" ]; then
    "$AGENT_PATH" collect | jq '.'
else
    echo "❌ Agent not found at $AGENT_PATH"
    echo "Run: cd agent && go build -o litemidgo-agent ."
fi
echo
echo

# Agent - Send Metrics to LiteMIDgo
echo "4. Agent - Send Metrics to LiteMIDgo:"
if [ -f "$AGENT_PATH" ]; then
    "$AGENT_PATH" send --server "$BASE_URL"
else
    echo "❌ Agent not found at $AGENT_PATH"
    echo "Run: cd agent && go build -o litemidgo-agent ."
fi
echo
echo

# Send Custom Endpoint Data
echo "5. Send Custom Endpoint Data:"
curl -s -X POST "$BASE_URL/proxy/ecc_queue" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "custom-agent",
    "topic": "endpointData",
    "name": "web-server-01",
    "source": "production",
    "payload": {
      "endpoint_type": "web_server",
      "status": "healthy",
      "response_time_ms": 145,
      "cpu_usage": 45.2,
      "memory_usage": 67.8,
      "active_connections": 1250,
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    }
  }' | jq '.'
echo
echo

# Send Application Metrics
echo "6. Send Application Metrics:"
curl -s -X POST "$BASE_URL/proxy/ecc_queue" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "app-monitor",
    "topic": "endpointData",
    "name": "api-gateway",
    "source": "k8s-cluster-01",
    "payload": {
      "endpoint_type": "api_gateway",
      "application": "user-service",
      "version": "2.1.3",
      "requests_per_second": 1250,
      "error_rate": 0.02,
      "p95_response_time": 230,
      "database_connections": 15,
      "cache_hit_rate": 94.5,
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    }
  }' | jq '.'
echo
echo

# Send Network Device Data
echo "7. Send Network Device Data:"
curl -s -X POST "$BASE_URL/proxy/ecc_queue" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "network-monitor",
    "topic": "endpointData",
    "name": "core-switch-01",
    "source": "datacenter-01",
    "payload": {
      "endpoint_type": "network_switch",
      "vendor": "Cisco",
      "model": "Catalyst 9300",
      "uptime_seconds": 1234567,
      "cpu_utilization": 23.5,
      "memory_utilization": 45.8,
      "interface_status": {
        "GigabitEthernet0/1": "up",
        "GigabitEthernet0/2": "up",
        "GigabitEthernet0/3": "down"
      },
      "temperature_celsius": 42.5,
      "power_consumption_watts": 125,
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    }
  }' | jq '.'
echo
echo

# Send Database Metrics
echo "8. Send Database Metrics:"
curl -s -X POST "$BASE_URL/proxy/ecc_queue" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": "db-monitor",
    "topic": "endpointData",
    "name": "postgres-primary",
    "source": "database-cluster-01",
    "payload": {
      "endpoint_type": "database",
      "engine": "postgresql",
      "version": "14.5",
      "connections": {
        "active": 45,
        "idle": 12,
        "max_allowed": 200
      },
      "performance": {
        "transactions_per_second": 1250,
        "query_duration_avg_ms": 15.2,
        "lock_wait_time_ms": 2.3
      },
      "storage": {
        "database_size_gb": 125.5,
        "table_count": 156,
        "index_size_gb": 45.2
      },
      "replication_lag_seconds": 0,
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    }
  }' | jq '.'
echo
echo

echo "Examples completed!"
echo
echo "Agent Usage:"
echo "  cd agent && go build -o litemidgo-agent ."
echo "  ./litemidgo-agent collect                    # Show metrics"
echo "  ./litemidgo-agent send --server $BASE_URL   # Send once"
echo "  ./litemidgo-agent daemon --server $BASE_URL # Run continuously"
echo
echo "Note: Make sure jq is installed for pretty JSON output"
