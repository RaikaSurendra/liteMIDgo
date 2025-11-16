#!/bin/bash

echo "🔍 LiteMIDgo Monitoring Script"
echo "================================"
echo "This script shows the status of your agent and server"
echo ""

# Check if server is running
echo "📡 Server Status:"
if lsof -i :8080 > /dev/null 2>&1; then
    echo "✅ LiteMIDgo server is running on port 8080"
    echo "   Process: $(lsof -i :8080 | tail -n 1 | awk '{print $1}') (PID: $(lsof -i :8080 | tail -n 1 | awk '{print $2}'))"
else
    echo "❌ LiteMIDgo server is not running"
fi

echo ""

# Check if agent is running
echo "🤖 Agent Status:"
AGENT_PID=$(pgrep -f "litemidgo-agent daemon")
if [ -n "$AGENT_PID" ]; then
    echo "✅ LiteMIDgo agent is running (PID: $AGENT_PID)"
    echo "   Sending metrics every 10 seconds"
else
    echo "❌ LiteMIDgo agent is not running"
fi

echo ""
echo "📊 Quick Commands:"
echo "  Start agent:     cd agent && ./litemidgo-agent daemon --interval 10"
echo "  Stop agent:      pkill -f 'litemidgo-agent daemon'"
echo "  Test single send: cd agent && ./litemidgo-agent send"
echo "  Test with debug: cd agent && ./litemidgo-agent send --debug"
echo "  View metrics:    cd agent && ./litemidgo-agent collect"
