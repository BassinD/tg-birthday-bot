# Telegram Birthday Bot 🥳

A serverless Telegram bot built in Go that manages birthday notifications for group chats. It utilizes Google Cloud Platform (Firestore & Cloud Run) and generates custom, AI-powered birthday wishes using the Gemini API.

## Architecture
* **Language:** Go 1.22+
* **Infrastructure:** GCP Cloud Run (Serverless via Docker)
* **Database:** GCP Firestore (Native mode)
* **AI:** Google Gemini API (`gemini-1.5-flash`)
* **Triggers:** Webhooks (for commands) and Cloud Scheduler (daily cron job)

## Features
* Works via Telegram webhooks (no long-polling).
* Saves birthdays by `@username` per chat.
* Configurable celebration times and AI prompt styles per chat.
* Zero-cost hosting utilizing GCP and Google AI Studio free tiers.

## Local Development Setup

1. **Clone the repository:**

```bash
   git clone [https://github.com/yourusername/tg-birthday-bot.git](https://github.com/yourusername/tg-birthday-bot.git)
   cd tg-birthday-bot

```

2. **Configure Environment Variables:**

Create a `.env` file in the root directory (this is ignored by Git) and add your credentials:

```env
TELEGRAM_BOT_TOKEN="your_telegram_token"
GEMINI_API_KEY="your_gemini_api_key"
GCP_PROJECT_ID="your_gcp_project_id"
PORT="8080"

```


3. **Run the bot:**
```bash
go run cmd/bot/main.go

```
