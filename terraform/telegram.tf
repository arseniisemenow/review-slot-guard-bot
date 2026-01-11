# Set Telegram webhook automatically after API Gateway is deployed
resource "telegram_bot_webhook" "rsgb_webhook" {
  url             = "${yandex_api_gateway.telegram_webhook.domain}/webhook"
}
