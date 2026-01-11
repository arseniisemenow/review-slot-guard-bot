variable "yandex_cloud_token" {
  description = "Yandex Cloud OAuth token (or use service account key)"
  type        = string
  sensitive   = true
}

variable "folder_id" {
  description = "Yandex Cloud folder ID"
  type        = string
}

variable "cloud_id" {
  description = "Yandex Cloud ID"
  type        = string
}

variable "telegram_bot_token" {
  description = "Telegram bot API token"
  type        = string
  sensitive   = true
}

variable "s21auto_api_token" {
  description = "s21auto API token (for reference - actual tokens stored per user in Lockbox)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "s21auto_api_url" {
  description = "s21auto API base URL"
  type        = string
  default     = "https://platform.21-school.ru/services/graphql"
}

variable "periodic_job_schedule" {
  description = "Cron expression for periodic job trigger"
  type        = string
  default     = "*/5 * * * *"
}

variable "function_memory" {
  description = "Memory allocated to each function (MB)"
  type        = number
  default     = 256
}

variable "function_timeout" {
  description = "Function execution timeout (seconds)"
  type        = number
  default     = 60
}
