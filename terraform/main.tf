terraform {
  required_version = ">= 1.5"
  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "~> 0.177"
    }
  }
}

provider "yandex" {
  zone     = "ru-central1-a"
  token    = var.yandex_cloud_token
}

# Note: Telegram webhook should be set manually after deployment
# Use: curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook" -d "url=<API_GATEWAY_URL>/webhook"
