# JobRadar

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

English | [简体中文](README_CN.md)

> Upwork Job Monitoring and Smart Notification Tool

JobRadar is a tool I built in Go to solve my own problem - finding the right Upwork jobs efficiently. It monitors Upwork RSS feeds for new job postings, filters them based on your criteria, and sends instant notifications via Telegram or Email.

## ✨ Features

- 🔍 **Smart Monitoring** - Monitors Upwork RSS feeds for new jobs
- 🎯 **Flexible Filtering** - Filter by budget, keywords, job type, and more
- 📱 **Instant Notifications** - Get notified via Telegram or Email
- ⏰ **Scheduled Checks** - Runs automatically at configurable intervals
- 🌙 **Quiet Hours** - Pause notifications during specified hours
- 🔄 **Deduplication** - Never see the same job twice
- 🐳 **Docker Ready** - Easy deployment with Docker

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- A Telegram Bot (for notifications)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/jobradar.git
cd jobradar

# Build
go build -o jobradar ./cmd/jobradar

# Or use make
make build
```

### Configuration

1. Copy the example configuration:

```bash
cp config.example.yaml config.yaml
```

2. **Get your Upwork RSS URL** (Important!):
   - Login to your Upwork account
   - Go to **Find Work** page
   - Set your search filters (keywords, budget, etc.)
   - Click the **RSS icon** (usually top-right of search results)
   - Copy the full URL (it contains your authentication token)

3. Edit `config.yaml` with your settings:

```yaml
name: "My Job Monitor"

# Use your authenticated RSS URLs from Upwork
rss_feeds:
  - name: "Golang Jobs"
    url: "https://www.upwork.com/ab/feed/jobs/rss?securityToken=YOUR_TOKEN&userUid=YOUR_UID&..."

filters:
  budget:
    min: 100
    max: 5000
  job_type: "fixed"
  max_proposals: 20
  exclude_keywords:
    - "lowest bid"
    - "cheap"

notifications:
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"

schedule:
  interval_minutes: 30
  quiet_hours:
    enabled: true
    start: "23:00"
    end: "07:00"
    timezone: "Asia/Shanghai"
```

4. Set environment variables:

```bash
export TELEGRAM_BOT_TOKEN="your_bot_token"
export TELEGRAM_CHAT_ID="your_chat_id"
```

> **Note**: Upwork no longer supports public RSS feeds. You must login to Upwork and get your personal RSS URL which includes authentication tokens.

### Usage

```bash
# Check for new jobs immediately
jobradar check

# Start scheduled monitoring
jobradar run

# View notification history
jobradar history

# View statistics
jobradar stats

# Validate configuration
jobradar validate

# Test notifications
jobradar test-notify
```

## 📱 Telegram Bot Setup

1. Open Telegram and search for `@BotFather`
2. Send `/newbot` and follow the prompts
3. Copy the bot token
4. Add the bot to a group or start a private chat
5. Get your chat ID:
   - Send a message to your bot
   - Visit `https://api.telegram.org/bot<TOKEN>/getUpdates`
   - Find the `chat.id` in the response

## 🐳 Docker Deployment

### Using Docker Compose

1. Create a `docker/config.yaml` file with your settings
2. Create a `.env` file:

```bash
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
```

3. Start the container:

```bash
cd docker
docker-compose up -d
```

### Using Docker directly

```bash
# Build
docker build -t jobradar -f docker/Dockerfile .

# Run
docker run -d \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -e TELEGRAM_BOT_TOKEN=xxx \
  -e TELEGRAM_CHAT_ID=xxx \
  jobradar
```

## 📊 Notification Format

When a matching job is found, you'll receive a notification like this:

```
🔔 New Job Match!

📋 Golang API Integration for E-commerce
💰 $300-500 (Fixed)
👥 Proposals: 5
⏰ Posted: 2 hours ago
🏷️ Skills: Golang, REST API, Microservices

📝 Looking for Go developer to build microservices...

🔗 View Job

---
✅ Matched: golang, api
```

## 🛠️ Development

### Project Structure

```
jobradar/
├── cmd/jobradar/         # Application entry point
├── cli/                  # CLI commands
├── internal/
│   ├── config/          # Configuration handling
│   ├── model/           # Data models
│   ├── fetcher/         # RSS fetching
│   ├── filter/          # Job filtering
│   ├── notifier/        # Notifications
│   ├── storage/         # SQLite storage
│   ├── scheduler/       # Cron scheduling
│   └── engine/          # Main engine
├── docker/              # Docker files
└── config.example.yaml  # Example configuration
```

### Building

```bash
# Build
make build

# Run tests
make test

# Format code
make fmt

# Run linter
make lint
```

## 📝 Configuration Reference

| Section | Option | Description | Default |
|---------|--------|-------------|---------|
| `searches` | `name` | Search configuration name | - |
| | `keywords` | Keywords to search for | - |
| `filters` | `budget.min` | Minimum budget | 0 |
| | `budget.max` | Maximum budget | 100000 |
| | `job_type` | fixed / hourly / all | all |
| | `posted_within_hours` | Max age of jobs | 24 |
| | `max_proposals` | Max proposal count | 20 |
| | `exclude_keywords` | Keywords to exclude | [] |
| `notifications` | `telegram.enabled` | Enable Telegram | false |
| | `email.enabled` | Enable Email | false |
| `schedule` | `interval_minutes` | Check interval | 30 |
| | `quiet_hours.enabled` | Enable quiet hours | false |
| `storage` | `database` | SQLite database path | jobradar.db |
| | `retention_days` | Days to keep records | 7 |

## 🎯 Why I Built This

As a freelancer on Upwork, I found myself constantly refreshing the job feed to catch new opportunities. This tool automates that process, allowing me to:

- Focus on my current work without missing new opportunities
- Get instant notifications for jobs that match my skills
- Filter out low-quality or unsuitable jobs automatically
- Track my job search statistics

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Built with ❤️ by a developer who got tired of refreshing Upwork manually.
