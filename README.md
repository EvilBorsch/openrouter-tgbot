# Telegram LLM Bot 🤖

A powerful Telegram bot that provides access to various Large Language Models (LLMs) through OpenRouter API. Built with Go for reliability and performance.

## 🌟 Features

- **Multiple LLM Support**: Access various models from OpenRouter (GPT-4, Claude, Gemini, etc.)
- **Chat Modes**: Switch between conversation modes with or without history
- **Precise Expense Tracking**: Real-time costs via OpenRouter generation stats API
- **Custom Models**: Add and manage your preferred models
- **Message Splitting**: Automatically handles long responses
- **User Authentication**: Restrict access to authorized users only
- **Data Persistence**: All settings and chat history are saved
- **Docker Support**: Easy deployment and scaling
- **Auto-Restart**: Graceful handling of crashes and restarts
- **Markdown Support**: Rich text formatting in responses

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ (for building from source)
- Docker (for containerized deployment)
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
- OpenRouter API Key (from [OpenRouter](https://openrouter.ai))

### Installation

#### Option 1: Using Make (Recommended)

```bash
# Clone the repository
git clone <your-repo-url>
cd telegrambot

# Setup environment and create config
make setup

# Edit config.json with your tokens
nano config.json

# Run the bot
make run
```

#### Option 2: Manual Setup

```bash
# Clone and build
git clone <your-repo-url>
cd telegrambot
go mod tidy
go build -o telegrambot .

# Create config file
./telegrambot  # This will create config.json

# Edit config and run
nano config.json
./telegrambot
```

#### Option 3: Docker

```bash
# Clone repository
git clone <your-repo-url>
cd telegrambot

# Create config
make setup
nano config.json

# Deploy with Docker
make docker-build
make docker-run
```

## ⚙️ Configuration

Edit `config.json` with your settings:

```json
{
  "telegram_token": "YOUR_BOT_TOKEN_FROM_BOTFATHER",
  "openrouter_api_key": "YOUR_OPENROUTER_API_KEY",
  "openrouter_base_url": "https://openrouter.ai/api/v1",
  "allowed_users": [123456789, 987654321],
  "default_model": "openai/gpt-3.5-turbo",
  "default_chat_mode": "without_history",
  "max_message_length": 4096,
  "log_level": "info",
  "data_directory": "data"
}
```

### Getting Required Tokens

1. **Telegram Bot Token**:
   - Message [@BotFather](https://t.me/BotFather) on Telegram
   - Create a new bot with `/newbot`
   - Copy the token

2. **OpenRouter API Key**:
   - Sign up at [OpenRouter](https://openrouter.ai)
   - Go to [API Keys](https://openrouter.ai/keys)
   - Create a new API key

3. **Your Telegram User ID**:
   - Start the bot (it will log unauthorized attempts)
   - Send a message to your bot
   - Check logs to find your user ID
   - Add it to `allowed_users` in config.json

## 📱 Usage

### Bot Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message and quick start guide |
| `/menu` | Show all available commands |
| `/mode [with_history\|without_history]` | Change chat mode |
| `/model [model_name]` | Change current LLM model |
| `/addmodel [model_name]` | Add a custom model |
| `/listmodels` | Show all available models |
| `/expenses` | Show usage statistics and costs |
| `/clear` | Clear chat history |
| `/status` | Show current settings |
| `/help` | Show help menu |

### Chat Modes

- **`without_history`** (default): Each message is independent
- **`with_history`**: AI remembers previous conversation context

### Popular Models

| Model | Description | Use Case |
|-------|-------------|----------|
| `openai/gpt-4` | Most capable, higher cost | Complex tasks, analysis |
| `openai/gpt-3.5-turbo` | Fast and affordable | General chat, simple tasks |
| `anthropic/claude-3-sonnet` | Great for analysis | Writing, reasoning |
| `google/gemini-pro` | Google's latest | Balanced performance |
| `mistralai/mistral-7b-instruct` | Open source | Cost-effective |

## 🔧 Development

### Available Make Commands

```bash
make help              # Show all available commands
make setup             # Install dependencies and create config
make build             # Build the application
make run               # Run the application
make dev               # Run with hot reload
make test              # Run tests
make clean             # Clean build artifacts

# Docker commands
make docker-build      # Build Docker image
make docker-run        # Run Docker container
make docker-stop       # Stop Docker container
make docker-logs       # Show container logs

# Production
make deploy            # Full deployment (build + run)
make backup            # Backup user data
make restore           # Restore from backup
make status            # Show bot status
make logs              # Show logs
```

### Project Structure

```
telegrambot/
├── main.go              # Application entry point
├── internal/
│   ├── bot/             # Telegram bot logic
│   │   ├── bot.go       # Core bot functionality
│   │   └── commands.go  # Command handlers
│   ├── config/          # Configuration management
│   │   └── config.go    # Config loading and validation
│   ├── openrouter/      # OpenRouter API client
│   │   └── client.go    # LLM API interactions
│   └── storage/         # Data persistence
│       └── storage.go   # File-based storage
├── data/               # User data (created automatically)
├── config.json         # Bot configuration
├── Dockerfile          # Container configuration
├── Makefile           # Build and deployment scripts
└── README.md          # This file
```

## 🚀 Deployment

### Local Development

```bash
# Quick start
make setup
make run

# With hot reload
make dev
```

### Production with Docker

```bash
# One-time setup
git clone <repo>
cd telegrambot
make setup
nano config.json  # Add your tokens

# Deploy
make deploy

# Monitor
make logs
make status
```

### Server Deployment

For production servers, use Docker with restart policies:

```bash
# Build and deploy
make docker-build

# Run with restart policy
docker run -d \
  --name telegrambot \
  --restart unless-stopped \
  -v $(PWD)/config.json:/app/config.json:ro \
  -v $(PWD)/data:/app/data \
  telegrambot:latest

# Check status
make status
make logs
```

### Systemd Service (Alternative to Docker)

Create `/etc/systemd/system/telegrambot.service`:

```ini
[Unit]
Description=Telegram LLM Bot
After=network.target

[Service]
Type=simple
User=botuser
WorkingDirectory=/opt/telegrambot
ExecStart=/opt/telegrambot/telegrambot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable telegrambot
sudo systemctl start telegrambot
```

## 📊 Monitoring

### Health Checks

The bot includes built-in health monitoring:
- Automatic restart on crashes
- Graceful shutdown handling
- Connection recovery
- Data persistence validation

### Logs

```bash
# Docker logs
make logs

# System logs (if using systemd)
sudo journalctl -u telegrambot -f

# Local development
./telegrambot  # Logs to stdout
```

### Backup and Recovery

```bash
# Create backup
make backup

# Restore from backup
make restore BACKUP=backups/backup-20231201-120000.tar.gz
```

## 🔒 Security

- **User Authentication**: Only configured users can access the bot
- **API Key Protection**: Keep your OpenRouter key secure
- **Container Security**: Runs as non-root user
- **Data Isolation**: User data stored separately
- **Input Validation**: All user inputs are validated

## 💰 Cost Management

The bot tracks all expenses with **precision and accuracy**:
- **Accurate OpenRouter API pricing** - Uses generation stats endpoint for real costs
- **Native token counting** - Model-specific tokenizers for precise usage
- Real-time cost tracking after each request
- Per-model usage statistics and comparisons
- Historical expense tracking with detailed breakdowns
- Weekly/monthly summaries

Check your usage with `/expenses` command to see exact costs and native token counts.

## 🐛 Troubleshooting

### Common Issues

1. **"Unauthorized user" error**
   - Check your user ID in logs
   - Add it to `allowed_users` in config.json

2. **"Failed to create bot API" error**
   - Verify your Telegram bot token
   - Ensure token is from @BotFather

3. **"OpenRouter API error" messages**
   - Check your OpenRouter API key
   - Verify you have credits in your account
   - Ensure model name is correct

4. **Bot not responding**
   - Check logs: `make logs`
   - Verify network connectivity
   - Restart: `make docker-stop && make docker-run`

### Debug Mode

Enable debug logging by setting `log_level: "debug"` in config.json.

### Getting Help

1. Check the logs first
2. Verify your configuration
3. Test with `/status` command
4. Use `/menu` to see available commands

## 📈 Performance

- **Concurrent Handling**: Multiple users supported simultaneously
- **Memory Efficient**: File-based storage with smart caching
- **Rate Limiting**: Built-in protection against API limits
- **Message Queuing**: Handles message bursts gracefully
- **Auto-scaling**: Docker containers can be replicated

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Update documentation
6. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🚀 Roadmap

- [ ] Web interface for configuration
- [ ] Voice message support
- [ ] Image generation capabilities  
- [ ] Multi-language support
- [ ] Advanced analytics dashboard
- [ ] Integration with more LLM providers

---

**Need help?** Open an issue or check the troubleshooting section above. 