output "api_gateway_url" {
  description = "Full URL for Telegram webhook registration"
  value       = "${yandex_api_gateway.telegram_webhook.domain}/webhook"
}

output "ydb_endpoint" {
  description = "YDB connection endpoint"
  value       = yandex_ydb_database_serverless.review_slot_guard_bot.ydb_endpoint
}

output "ydb_database" {
  description = "YDB database name"
  value       = yandex_ydb_database_serverless.review_slot_guard_bot.database
}

output "function_periodic_job_id" {
  description = "PeriodicJob function ID"
  value       = yandex_function.periodic_job.id
}

output "function_telegram_handler_id" {
  description = "TelegramHandler function ID"
  value       = yandex_function.telegram_handler.id
}

output "lockbox_secret_id" {
  description = "Lockbox secret ID"
  value       = yandex_lockbox_secret.review_slot_guard_bot.id
}

output "service_account_id" {
  description = "Service account ID"
  value       = yandex_iam_service_account.review_slot_guard_bot.id
}
