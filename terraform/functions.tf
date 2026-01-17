# PeriodicJob function
data "archive_file" "periodic_job" {
  type        = "zip"
  source_dir  = "${path.module}/../build/periodic_job"
  output_path = "${path.module}/periodic_job.zip"
}

resource "yandex_function" "periodic_job" {
  name        = "rsgb-periodic-job"
  description = "Periodic job: Slot optimization and review processing"
  folder_id   = var.folder_id
  user_hash   = data.archive_file.periodic_job.output_sha256
  runtime     = "golang123"
  entrypoint  = "main.Handler"
  memory      = var.function_memory

  content {
    zip_filename = data.archive_file.periodic_job.output_path
  }

  environment = {
    YDB_ENDPOINT      = "grpcs://ydb.serverless.yandexcloud.net:2135"
    YDB_DATABASE      = yandex_ydb_database_serverless.review_slot_guard_bot.database_path
  }

  service_account_id = yandex_iam_service_account.review_slot_guard_bot.id
}

# TelegramHandler function
data "archive_file" "telegram_handler" {
  type        = "zip"
  source_dir  = "${path.module}/../build/telegram_handler"
  output_path = "${path.module}/telegram_handler.zip"
}

resource "yandex_function" "telegram_handler" {
  name        = "rsgb-telegram-handler"
  description = "Telegram webhook handler: User button clicks and commands"
  folder_id   = var.folder_id
  user_hash   = data.archive_file.telegram_handler.output_sha256
  runtime     = "golang123"
  entrypoint  = "main.Handler"
  memory      = var.function_memory

  content {
    zip_filename = data.archive_file.telegram_handler.output_path
  }

  environment = {
    YDB_ENDPOINT         = "grpcs://ydb.serverless.yandexcloud.net:2135"
    YDB_DATABASE         = yandex_ydb_database_serverless.review_slot_guard_bot.database_path
    TELEGRAM_BOT_TOKEN    = var.telegram_bot_token
  }

  service_account_id = yandex_iam_service_account.review_slot_guard_bot.id
}
