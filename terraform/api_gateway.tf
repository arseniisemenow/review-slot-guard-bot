# API Gateway for Telegram webhook
resource "yandex_api_gateway" "telegram_webhook" {
  folder_id   = var.folder_id
  name        = "rsgb-webhook"
  description = "Telegram webhook endpoint for TelegramHandler"

  spec = templatefile("${path.module}/openapi.yaml", {
    function_id = yandex_function.telegram_handler.id
    folder_id   = var.folder_id
  })
}
