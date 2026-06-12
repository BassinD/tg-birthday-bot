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
   git clone [https://github.com/BassinD/tg-birthday-bot.git](https://github.com/BassinD/tg-birthday-bot.git)
   cd tg-birthday-bot

```

2. **Configure Environment Variables:**

Create a `.env` file in the root directory (this is ignored by Git) and add your credentials:

```env
TELEGRAM_BOT_TOKEN="your_telegram_token"
GEMINI_API_KEY="your_gemini_api_key"
GCP_PROJECT_ID="your_gcp_project_id"
FIRESTORE_DB_ID="firestore_db_id"
PORT="8080"

```


3. **Run the bot:**
```bash
go run cmd/bot/main.go

```

## 🚀 Deployment (Google Cloud)

This project is designed to run serverless on Google Cloud Run, utilizing Artifact Registry for image storage, Secret Manager for API keys, and Cloud Scheduler for automated daily triggers.

**Prerequisites:**

* Ensure the Google Cloud CLI (`gcloud`) is installed and authenticated.
* Ensure you have created your Secrets (`TELEGRAM_BOT_TOKEN`, `GEMINI_API_KEY`) and granted the Compute Service Account access to them.
* Replace `YOUR_PROJECT_ID`, `YOUR_REPO_NAME`, and `YOUR_CLOUD_RUN_URL` with your actual Google Cloud environment details before executing.

### 1. Build and Push the Docker Image

Build the Go application in the cloud and push the container to Artifact Registry (`europe-west3`):

```powershell
gcloud builds submit --tag europe-west3-docker.pkg.dev/YOUR_PROJECT_ID/YOUR_REPO_NAME/birthday-bot:v1

```

### 2. Deploy to Cloud Run

Deploy the container, inject the secure API keys from Secret Manager, and generate the public web service URL:

```powershell
gcloud run deploy birthday-bot `
    --image europe-west3-docker.pkg.dev/YOUR_PROJECT_ID/YOUR_REPO_NAME/birthday-bot:v1 `
    --region europe-west3 `
    --allow-unauthenticated `
    --set-env-vars="GCP_PROJECT_ID=YOUR_PROJECT_ID,PORT=8080,FIRESTORE_DB_ID=firestore_db_id" `
    --update-secrets="TELEGRAM_BOT_TOKEN=TELEGRAM_BOT_TOKEN:latest,GEMINI_API_KEY=GEMINI_API_KEY:latest"

```

*(Note: Once deployed, copy the generated Service URL and register it with the Telegram API using the `setWebhook` endpoint).*

### 3. Create the Automated Cron Job

Set up Cloud Scheduler to ping the application's `/cron` endpoint at the top of every hour to check for matching local timezones and dispatch AI greetings:

```powershell
gcloud scheduler jobs create http birthday-bot-trigger `
    --location="europe-west3" `
    --schedule="0 * * * *" `
    --time-zone="UTC" `
    --uri="YOUR_CLOUD_RUN_URL/cron" `
    --http-method="GET" `
    --description="Hourly trigger for Telegram Birthday Bot"

```

### 4. Trigger the Cron Job Manually

Force the scheduler to execute immediately to test database reads and AI generation without waiting for the top of the hour:

```powershell
gcloud scheduler jobs run birthday-bot-trigger --location="europe-west3"

```
