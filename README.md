# Review Slot Guard Bot

A Telegram bot for School 21 reviewers that automates review slot management and intelligently handles review requests based on project whitelists.

## Features

- **Automatic Slot Monitoring**: Continuously monitors your School 21 calendar for new review bookings
- **Smart Whitelist Management**: Auto-approves reviews from whitelisted projects/families
- **Intelligent Slot Shifting**: Automatically shifts whitelisted review slots closer to current time
- **Telegram Integration**: Receive notifications and approve/decline reviews directly in Telegram
- **Configurable Settings**: Fine-tune behavior with customizable thresholds and timeouts

## Architecture

The bot consists of two Yandex Cloud Functions:

1. **telegram_handler**: Handles incoming Telegram messages, commands, and button callbacks
2. **periodic_job**: Runs every 5 minutes to process review requests and calendar events

### State Machine

Review requests progress through these states:

```
UNKNOWN_PROJECT_REVIEW
    -> (extract project from notification)
KNOWN_PROJECT_REVIEW
    -> (whitelisted) -> WHITELISTED
    -> (not whitelisted) -> NOT_WHITELISTED
    -> (deadline approaching) -> NEED_TO_APPROVE
WHITELISTED -> (shift slot if needed) -> APPROVED
NOT_WHITELISTED -> (timeout) -> AUTO_CANCELLED_NOT_WHITELISTED
NEED_TO_APPROVE -> (send Telegram message) -> WAITING_FOR_APPROVE
WAITING_FOR_APPROVE
    -> (user approves) -> APPROVED
    -> (user declines) -> CANCELLED
    -> (timeout) -> AUTO_CANCELLED
```

## Prerequisites

- Go 1.23+
- Yandex Cloud account
- Telegram bot (create via [@BotFather](https://t.me/botfather))
- School 21 account

## Environment Variables

### Required for Yandex Cloud Functions

| Variable | Description | Source |
|----------|-------------|--------|
| `YDB_ENDPOINT` | YDB database endpoint | Terraform output |
| `YDB_DATABASE` | YDB database name | Terraform output |
| `LOCKBOX_SECRET_ID` | Lockbox secret ID | Terraform output |
| `TELEGRAM_BOT_TOKEN` | Telegram bot API token | From @BotFather |

### Required for Local Testing

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `YDB_ENDPOINT` | YDB endpoint | `grpcs://ydb.serverless.yandexcloud.net:2135` |
| `YDB_DATABASE` | Database name | `/ru-central1/b1xxx/dbxxx` |
| `LOCKBOX_SECRET_ID` | Secret ID | `e6exxxxx` |

### Terraform Variables

Create `terraform.tfvars` file:

```hcl
folder_id            = "your-folder-id"
cloud_id             = "your-cloud-id"
telegram_bot_token   = "your-telegram-bot-token"
s21auto_api_token    = ""  # Reference only - user tokens stored per-user
s21auto_api_url      = "https://platform.21-school.ru/services/graphql"
periodic_job_schedule = "*/5 * * * *"
function_memory      = 256
function_timeout     = 60
```

## Deployment

### 1. Configure Terraform

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
```

### 2. Initialize and Apply

```bash
terraform init
terraform plan
terraform apply
```

### 3. Build and Deploy Functions

```bash
# From project root
cd functions/telegram_handler
GOOS=linux GOARCH=amd64 go build -o main .
cd ../periodic_job
GOOS=linux GOARCH=amd64 go build -o main .
cd ../..
terraform apply  # Re-apply to upload new zips
```

### 4. Set Telegram Webhook

The webhook is automatically configured via Terraform using the API Gateway URL.

## Local Development

### Run Telegram Handler Locally

```bash
cd functions/telegram_handler
export YDB_ENDPOINT="grpcs://..."
export YDB_DATABASE="/ru-central1/..."
export LOCKBOX_SECRET_ID="xxx"
export TELEGRAM_BOT_TOKEN="xxx"
go run main.go
# Server runs on :8080
```

### Run Periodic Job Locally

```bash
cd functions/periodic_job
export YDB_ENDPOINT="grpcs://..."
export YDB_DATABASE="/ru-central1/..."
export LOCKBOX_SECRET_ID="xxx"
go run main.go
```

## Telegram Commands

| Command | Description |
|---------|-------------|
| `/start` | Start authentication flow |
| `/logout` | Log out and clear credentials |
| `/status` | Show current status and active reviews |
| `/settings` | Display current settings |
| `/whitelist` | Show whitelisted projects and families |
| `/whitelist_add <family|project> <name>` | Add to whitelist |
| `/whitelist_remove <name>` | Remove from whitelist |
| `/set_deadline_shift <minutes>` | Response deadline shift (1-60) |
| `/set_cancel_delay <minutes>` | Non-whitelist cancel delay (1-10) |
| `/set_slot_shift_threshold <minutes>` | Slot shift threshold (5-60) |
| `/set_slot_shift_duration <minutes>` | Slot shift duration (5-60) |
| `/set_cleanup_duration <minutes>` | Cleanup duration (15, 30, 45, 60) |
| `/set_notify_whitelist_timeout <true|false>` | Notify on whitelist timeout |
| `/set_notify_non_whitelist_cancel <true|false>` | Notify on non-whitelist cancel |
| `/help` | Show help message |

## Authentication Flow

1. User sends `/start` to bot
2. Bot requests credentials in format `login:password`
3. Bot authenticates with School 21 API
4. Access/refresh tokens stored in Lockbox
5. User record created in YDB

## Project Structure

```
review-slot-guard-bot/
├── common/
│   └── pkg/
│       ├── external/       # School 21 API client
│       ├── lockbox/        # Yandex Lockbox integration
│       ├── models/         # Data models
│       ├── telegram/       # Telegram bot client
│       ├── timeutil/       # Time utilities
│       └── ydb/            # YDB client and repository
├── functions/
│   ├── periodic_job/       # Background processing function
│   │   └── internal/logic/ # Business logic
│   └── telegram_handler/   # Telegram webhook handler
│       └── internal/handlers/
│           ├── callbacks.go # Button handlers
│           └── commands.go  # Command handlers
└── terraform/              # Infrastructure as Code
```

## Database Schema

### users
| Column | Type |
|--------|------|
| reviewer_login | Utf8 (PK) |
| status | Utf8 |
| telegram_chat_id | Int64 |
| created_at | Datetime |
| last_auth_success_at | Datetime |
| last_auth_failure_at | Datetime |

### user_settings
| Column | Type |
|--------|------|
| reviewer_login | Utf8 (PK) |
| response_deadline_shift_minutes | Int32 |
| non_whitelist_cancel_delay_minutes | Int32 |
| notify_whitelist_timeout | Bool |
| notify_non_whitelist_cancel | Bool |
| slot_shift_threshold_minutes | Int32 |
| slot_shift_duration_minutes | Int32 |
| cleanup_durations_minutes | Int32 |

### user_project_whitelist
| Column | Type |
|--------|------|
| reviewer_login | Utf8 (PK) |
| entry_type | Utf8 (PK) |
| name | Utf8 (PK) |

### project_families
| Column | Type |
|--------|------|
| family_label | Utf8 (PK) |
| project_name | Utf8 (PK) |

### review_requests
| Column | Type |
|--------|------|
| id | Utf8 (PK) |
| reviewer_login | Utf8 |
| notification_id | Utf8 |
| project_name | Utf8 |
| family_label | Utf8 |
| review_start_time | Datetime |
| calendar_slot_id | Utf8 |
| decision_deadline | Datetime |
| non_whitelist_cancel_at | Datetime |
| telegram_message_id | Utf8 |
| status | Utf8 |
| created_at | Datetime |
| decided_at | Datetime |

## License

MIT
