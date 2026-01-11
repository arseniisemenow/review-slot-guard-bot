# Lockbox Secret for storing user tokens and bot credentials
resource "yandex_lockbox_secret" "review_slot_guard_bot" {
  folder_id = var.folder_id
  name      = "rsgb-secrets"
}

# Secret Entries - Bot credentials and user tokens
resource "yandex_lockbox_secret_version" "review_slot_guard_bot" {
  secret_id = yandex_lockbox_secret.review_slot_guard_bot.id
  entries {
    key        = "telegram-bot-token"
    text_value = var.telegram_bot_token
  }
  entries {
    key        = "s21auto-api-url"
    text_value = var.s21auto_api_url
  }
  entries {
    key        = "users"
    text_value = jsonencode({
      "version": 1,
      "users"   = {}
    })
  }
}
