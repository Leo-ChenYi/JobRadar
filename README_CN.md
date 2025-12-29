# JobRadar

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[English](README.md) | 简体中文

> Upwork 工作机会监控与智能推送工具

JobRadar 是我用 Go 语言开发的一个工具，用于解决自己的实际问题——高效地发现 Upwork 上的合适工作。它能监控 Upwork RSS 订阅源，根据你设定的条件筛选工作，并通过 Telegram 或邮件即时推送通知。

## ✨ 功能特性

- 🔍 **智能监控** - 监控 Upwork RSS 订阅源获取新工作
- 🎯 **灵活筛选** - 按预算、关键词、项目类型等多维度筛选
- 📱 **即时通知** - 通过 Telegram 或邮件接收通知
- ⏰ **定时检查** - 可配置的自动定时检查
- 🌙 **安静时段** - 指定时段内暂停通知
- 🔄 **智能去重** - 同一工作不会重复推送
- 🐳 **Docker 支持** - 支持 Docker 一键部署

## 🚀 快速开始

### 环境要求

- Go 1.21 或更高版本
- Telegram Bot（用于接收通知）

### 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/jobradar.git
cd jobradar

# 构建
go build -o jobradar ./cmd/jobradar

# 或使用 make
make build
```

### 配置

1. 复制示例配置文件：

```bash
cp config.example.yaml config.yaml
```

2. **获取你的 Upwork RSS URL**（重要！）：
   - 登录你的 Upwork 账户
   - 进入 **Find Work** 页面
   - 设置搜索条件（关键词、预算等）
   - 点击搜索结果右上角的 **RSS 图标**
   - 复制完整的 URL（包含认证 token）

3. 编辑 `config.yaml` 配置你的设置：

```yaml
name: "我的工作监控"

# 使用从 Upwork 获取的认证 RSS URL
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

4. 设置环境变量：

```bash
export TELEGRAM_BOT_TOKEN="你的机器人token"
export TELEGRAM_CHAT_ID="你的聊天ID"
```

> **注意**：Upwork 不再支持公开的 RSS 订阅源。你必须登录 Upwork 并获取包含认证 token 的个人 RSS URL。

### 使用方法

```bash
# 立即检查新工作
jobradar check

# 启动定时监控
jobradar run

# 查看通知历史
jobradar history

# 查看统计信息
jobradar stats

# 验证配置文件
jobradar validate

# 测试通知功能
jobradar test-notify
```

## 📱 Telegram Bot 设置

1. 在 Telegram 中搜索 `@BotFather`
2. 发送 `/newbot` 并按提示操作
3. 复制获得的 Bot Token
4. 将 Bot 添加到群组或开始私聊
5. 获取 Chat ID：
   - 给你的 Bot 发送一条消息
   - 访问 `https://api.telegram.org/bot<TOKEN>/getUpdates`
   - 在返回结果中找到 `chat.id`

## 🐳 Docker 部署

### 使用 Docker Compose

1. 在 `docker/` 目录下创建 `config.yaml` 配置文件
2. 创建 `.env` 文件：

```bash
TELEGRAM_BOT_TOKEN=你的token
TELEGRAM_CHAT_ID=你的聊天ID
```

3. 启动容器：

```bash
cd docker
docker-compose up -d
```

### 直接使用 Docker

```bash
# 构建镜像
docker build -t jobradar -f docker/Dockerfile .

# 运行容器
docker run -d \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -e TELEGRAM_BOT_TOKEN=xxx \
  -e TELEGRAM_CHAT_ID=xxx \
  jobradar
```

## 📊 通知格式

当匹配到工作时，你会收到如下格式的通知：

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

## 🛠️ 开发指南

### 项目结构

```
jobradar/
├── cmd/jobradar/         # 程序入口
├── cli/                  # CLI 命令
├── internal/
│   ├── config/          # 配置管理
│   ├── model/           # 数据模型
│   ├── fetcher/         # RSS 获取
│   ├── filter/          # 工作筛选
│   ├── notifier/        # 通知推送
│   ├── storage/         # SQLite 存储
│   ├── scheduler/       # 定时调度
│   └── engine/          # 主引擎
├── docker/              # Docker 配置
└── config.example.yaml  # 配置示例
```

### 构建命令

```bash
# 构建
make build

# 运行测试
make test

# 格式化代码
make fmt

# 运行 linter
make lint
```

## 📝 配置参考

| 配置项 | 选项 | 说明 | 默认值 |
|--------|------|------|--------|
| `searches` | `name` | 搜索配置名称 | - |
| | `keywords` | 搜索关键词 | - |
| `filters` | `budget.min` | 最低预算 | 0 |
| | `budget.max` | 最高预算 | 100000 |
| | `job_type` | fixed / hourly / all | all |
| | `posted_within_hours` | 工作发布时间限制 | 24 |
| | `max_proposals` | 最大投标人数 | 20 |
| | `exclude_keywords` | 排除关键词 | [] |
| `notifications` | `telegram.enabled` | 启用 Telegram | false |
| | `email.enabled` | 启用邮件 | false |
| `schedule` | `interval_minutes` | 检查间隔（分钟） | 30 |
| | `quiet_hours.enabled` | 启用安静时段 | false |
| `storage` | `database` | SQLite 数据库路径 | jobradar.db |
| | `retention_days` | 记录保留天数 | 7 |

## 🎯 为什么开发这个工具

作为 Upwork 上的自由职业者，我发现自己需要不断刷新工作列表来发现新机会。这个工具自动化了这个过程，让我能够：

- 专注于当前工作，同时不错过新机会
- 第一时间收到符合技能的工作通知
- 自动过滤低质量或不合适的工作
- 追踪求职统计数据

## 🤝 贡献

欢迎贡献代码！请随时提交 Pull Request。

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

---

由一个厌倦了手动刷新 Upwork 的开发者用 ❤️ 构建。

