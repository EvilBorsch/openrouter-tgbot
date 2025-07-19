# Variables
BINARY_NAME=telegrambot
DOCKER_IMAGE=telegrambot
DOCKER_TAG=latest
CONFIG_FILE=config.json

# Default target
.PHONY: help
help: ## Display this help message
	@echo "Telegram LLM Bot - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start:"
	@echo "  1. make setup          - Install dependencies and create config"
	@echo "  2. Edit config.json    - Add your tokens and allowed users"
	@echo "  3. make run            - Start the bot"

# Development
.PHONY: setup
setup: ## Install dependencies and create default config
	@echo "Setting up development environment..."
	go mod tidy
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "Creating default config file..."; \
		go run . 2>/dev/null || true; \
		echo ""; \
		echo "✅ Config file created: $(CONFIG_FILE)"; \
		echo "📝 Please edit config.json and add:"; \
		echo "   - telegram_token: Your bot token from @BotFather"; \
		echo "   - openrouter_api_key: Your OpenRouter API key"; \
		echo "   - allowed_users: Array of allowed Telegram user IDs"; \
	else \
		echo "Config file already exists: $(CONFIG_FILE)"; \
	fi

.PHONY: build
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "✅ Build complete: $(BINARY_NAME)"

.PHONY: run
run: ## Run the application
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "❌ Config file not found. Run 'make setup' first."; \
		exit 1; \
	fi
	@echo "Starting $(BINARY_NAME)..."
	go run .

.PHONY: dev
dev: ## Run in development mode with auto-restart
	@if ! command -v air > /dev/null; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "✅ Cleaned"

# Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "❌ Config file not found. Run 'make setup' first."; \
		exit 1; \
	fi
	@echo "Preparing data directory..."
	@mkdir -p data
	@echo "Starting Docker container..."
	docker run -d \
		--name $(BINARY_NAME) \
		--restart unless-stopped \
		-v $(PWD)/config.json:/app/config.json:ro \
		-v $(PWD)/data:/app/data \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "✅ Container started: $(BINARY_NAME)"

.PHONY: docker-stop
docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME) || true
	docker rm $(BINARY_NAME) || true
	@echo "✅ Container stopped"

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	docker logs -f $(BINARY_NAME)

.PHONY: docker-shell
docker-shell: ## Access Docker container shell
	docker exec -it $(BINARY_NAME) /bin/sh

# Production deployment
.PHONY: deploy
deploy: docker-build docker-stop docker-run ## Deploy to production (build, stop old, start new)
	@echo "✅ Deployment complete"

.PHONY: backup
backup: ## Backup user data
	@echo "Creating backup..."
	@mkdir -p backups
	@tar -czf backups/backup-$$(date +%Y%m%d-%H%M%S).tar.gz data/
	@echo "✅ Backup created in backups/"

.PHONY: restore
restore: ## Restore from backup (usage: make restore BACKUP=backups/backup-xxx.tar.gz)
	@if [ -z "$(BACKUP)" ]; then \
		echo "❌ Please specify backup file: make restore BACKUP=backups/backup-xxx.tar.gz"; \
		exit 1; \
	fi
	@echo "Restoring from $(BACKUP)..."
	@tar -xzf $(BACKUP)
	@echo "✅ Restore complete"

# Monitoring
.PHONY: status
status: ## Show bot status
	@if docker ps | grep -q $(BINARY_NAME); then \
		echo "✅ Bot is running"; \
		docker ps | grep $(BINARY_NAME); \
	else \
		echo "❌ Bot is not running"; \
	fi

.PHONY: logs
logs: ## Show application logs
	@if docker ps | grep -q $(BINARY_NAME); then \
		docker logs --tail 100 -f $(BINARY_NAME); \
	else \
		echo "❌ Bot is not running"; \
	fi

# Utilities
.PHONY: fix-permissions
fix-permissions: ## Fix data directory permissions (if having Docker issues)
	@echo "Fixing data directory permissions..."
	@mkdir -p data
	@sudo chown -R $(shell id -u):$(shell id -g) data/
	@chmod -R 755 data/
	@echo "✅ Permissions fixed"

.PHONY: config-check
config-check: ## Validate configuration file
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "❌ Config file not found: $(CONFIG_FILE)"; \
		exit 1; \
	fi
	@echo "Validating config file..."
	@go run . --check-config 2>/dev/null || echo "⚠️  Config validation failed - please check your settings"

.PHONY: user-id
user-id: ## Get your Telegram user ID (send /start to bot first)
	@echo "To get your Telegram user ID:"
	@echo "1. Start your bot"
	@echo "2. Send any message to the bot"
	@echo "3. Check the logs - your user ID will be displayed"
	@echo "4. Add your user ID to allowed_users in config.json"

.PHONY: update
update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
	@echo "✅ Dependencies updated" 