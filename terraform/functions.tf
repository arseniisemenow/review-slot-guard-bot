# Archive PeriodicJob function
data "archive_file" "periodic_job" {
  type        = "zip"
  source_dir  = "${path.module}/../functions/periodic_job"
  output_path = "${path.module}/periodic_job.zip"
}

# Archive TelegramHandler function
data "archive_file" "telegram_handler" {
  type        = "zip"
  source_dir  = "${path.module}/../functions/telegram_handler"
  output_path = "${path.module}/telegram_handler.zip"
}

# PeriodicJob function
resource "yandex_function" "periodic_job" {
  name        = "rsgb-periodic-job"
  description = "Periodic job: Slot optimization and review processing"
  user_hash   = data.archive_file.periodic_job.output_sha256
  runtime     = "golang123"
  entrypoint  = "main.Handler"
  memory      = var.function_memory

  content {
    zip_filename = data.archive_file.periodic_job.output_path
  }

  environment = {
    YDB_ENDPOINT      = yandex_ydb_database_serverless.review_slot_guard_bot.ydb_endpoint
    YDB_DATABASE      = yandex_ydb_database_serverless.review_slot_guard_bot.database
    LOCKBOX_SECRET_ID = yandex_lockbox_secret.review_slot_guard_bot.id
  }

  service_account_id = yandex_iam_service_account.review_slot_guard_bot.id
}

# TelegramHandler function
resource "yandex_function" "telegram_handler" {
  name        = "rsgb-telegram-handler"
  description = "Telegram webhook handler: User button clicks and commands"
  user_hash   = data.archive_file.telegram_handler.output_sha256
  runtime     = "golang123"
  entrypoint  = "main.Handler"
  memory      = var.function_memory

  content {
    zip_filename = data.archive_file.telegram_handler.output_path
  }

  environment = {
    YDB_ENDPOINT      = yandex_ydb_database_serverless.review_slot_guard_bot.ydb_endpoint
    YDB_DATABASE      = yandex_ydb_database_serverless.review_slot_guard_bot.database
    LOCKBOX_SECRET_ID = yandex_lockbox_secret.review_slot_guard_bot.id
  }

  service_account_id = yandex_iam_service_account.review_slot_guard_bot.id
}
