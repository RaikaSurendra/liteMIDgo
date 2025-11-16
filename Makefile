# LiteMIDgo Makefile

.PHONY: help build server agent clean test restart stop status docker-build docker-up docker-down docker-logs docker-clean

# Default target
help: ## Show this help message
	@echo "LiteMIDgo - Lightweight ServiceNow MID Server Alternative"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build both server and agent
	@echo "🔨 Building LiteMIDgo server..."
	go build -o litemidgo .
	@echo "🔨 Building LiteMIDgo agent..."
	cd agent && go build -o litemidgo-agent .
	@echo "✅ Build complete!"

build-server: ## Build only the server
	@echo "🔨 Building LiteMIDgo server..."
	go build -o litemidgo .
	@echo "✅ Server build complete!"

build-agent: ## Build only the agent
	@echo "🔨 Building LiteMIDgo agent..."
	cd agent && go build -o litemidgo-agent .
	@echo "✅ Agent build complete!"

# Run targets
server: ## Run the server (builds first if needed)
	@if [ ! -f ./litemidgo ]; then $(MAKE) build-server; fi
	@echo "🚀 Starting LiteMIDgo server..."
	./litemidgo server-simple

agent: ## Run the agent (builds first if needed)
	@if [ ! -f ./agent/litemidgo-agent ]; then $(MAKE) build-agent; fi
	@echo "🚀 Starting LiteMIDgo agent..."
	cd agent && ./litemidgo-agent daemon --interval 10

debug-server: ## Run server with debug mode
	@if [ ! -f ./litemidgo ]; then $(MAKE) build-server; fi
	@echo "🐛 Starting LiteMIDgo server in debug mode..."
	./litemidgo server-simple --debug

debug-agent: ## Run agent with debug mode
	@if [ ! -f ./agent/litemidgo-agent ]; then $(MAKE) build-agent; fi
	@echo "🐛 Starting LiteMIDgo agent in debug mode..."
	cd agent && ./litemidgo-agent daemon --debug --interval 10

# Management targets
start: ## Start both server and agent in background
	@echo "🚀 Starting LiteMIDgo services..."
	$(MAKE) stop > /dev/null 2>&1 || true
	@if [ ! -f ./litemidgo ]; then $(MAKE) build-server > /dev/null; fi
	@if [ ! -f ./agent/litemidgo-agent ]; then $(MAKE) build-agent > /dev/null; fi
	./litemidgo server-simple > server.log 2>&1 &
	@echo "✅ Server started (PID: $$!)"
	cd agent && ./litemidgo-agent daemon --interval 10 > agent.log 2>&1 &
	@echo "✅ Agent started (PID: $$!)"
	@echo "📋 Use 'make status' to check running services"

stop: ## Stop all running services
	@echo "🛑 Stopping LiteMIDgo services..."
	@pkill -f "litemidgo server" 2>/dev/null || echo "No server running"
	@pkill -f "litemidgo-agent daemon" 2>/dev/null || echo "No agent running"
	@echo "✅ Services stopped"

restart: ## Restart both services
	@echo "🔄 Restarting LiteMIDgo services..."
	$(MAKE) stop
	sleep 2
	$(MAKE) start

status: ## Show status of running services
	@echo "📊 LiteMIDgo Service Status:"
	@echo ""
	@ps aux | grep -E "(liteMIDgo|litemidgo)" | grep -v grep || echo "No services running"

# Utility targets
logs: ## Show logs from background services
	@echo "📋 Server Logs:"
	@if [ -f server.log ]; then tail -20 server.log; else echo "No server log file"; fi
	@echo ""
	@echo "📋 Agent Logs:"
	@if [ -f agent/agent.log ]; then tail -20 agent/agent.log; else echo "No agent log file"; fi

test: ## Run tests
	@echo "🧪 Running tests..."
	go test ./...

test-config: ## Test configuration and ServiceNow connection
	@echo "🔧 Testing configuration..."
	./litemidgo config test

config: ## Run interactive configuration setup
	@echo "⚙️ Starting configuration setup..."
	./litemidgo config

collect: ## Collect and display system metrics (agent)
	@echo "📊 Collecting system metrics..."
	cd agent && ./litemidgo-agent collect

# Docker targets
docker-build: ## Build Docker images
	@echo "🐳 Building Docker images..."
	docker-compose build

docker-up: ## Start services with Docker Compose
	@echo "🐳 Starting LiteMIDgo services with Docker..."
	@if [ ! -f .env ]; then \
		echo "📋 Creating .env file from template..."; \
		cp .env.docker .env; \
		echo "⚠️  Please update .env with your actual ServiceNow credentials"; \
	fi
	docker-compose up -d

docker-down: ## Stop Docker services
	@echo "🐳 Stopping Docker services..."
	docker-compose down

docker-logs: ## Show Docker service logs
	@echo "📋 Docker service logs:"
	docker-compose logs -f

docker-status: ## Show Docker service status
	@echo "📊 Docker Service Status:"
	docker-compose ps

docker-clean: ## Clean Docker images and containers
	@echo "🧹 Cleaning up Docker..."
	docker-compose down --rmi all --volumes --remove-orphans
	docker system prune -f

docker-restart: ## Restart Docker services
	@echo "🔄 Restarting Docker services..."
	$(MAKE) docker-down
	sleep 2
	$(MAKE) docker-up

# Development targets
dev: ## Start services in development mode with auto-restart
	@echo "🔧 Starting development mode..."
	@echo "Server and agent will restart automatically on file changes"
	@echo "Press Ctrl+C to stop"
	@while true; do $(MAKE) restart; sleep 30; done

clean: ## Clean build artifacts and logs
	@echo "🧹 Cleaning up..."
	rm -f litemidgo agent/litemidgo-agent
	rm -f server.log agent.log
	rm -f *.log
	@echo "✅ Clean complete!"

install-deps: ## Install Go dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	cd agent && go mod download
	@echo "✅ Dependencies installed!"

# Quick start targets
quick-start: ## Build and start both services
	@echo "🚀 Quick starting LiteMIDgo..."
	$(MAKE) build
	$(MAKE) start

quick-stop: ## Stop services and clean up
	@echo "🛑 Quick stopping LiteMIDgo..."
	$(MAKE) stop
	$(MAKE) clean
