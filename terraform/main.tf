terraform {
  required_version = ">= 1.5"
  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "~> 0.177"
    }
    telegram = {
      source  = "yi-jiayu/telegram"
      version = "~> 0.3"
    }
  }
  backend "s3" {
    # State storage configuration - should be configured per deployment
    # bucket     = "your-state-bucket"
    # key        = "review-slot-guard-bot/terraform.tfstate"
    # region     = "ru-central1"
    # endpoints  = ["storage.yandexcloud.net"]
  }
}

provider "yandex" {
  zone = "ru-central1-a"
}

provider "telegram" {
  token = var.telegram_bot_token
}
