# YDB Serverless Database
resource "yandex_ydb_database_serverless" "review_slot_guard_bot" {
  folder_id           = var.folder_id
  name                = "rsgb-ydb"
  location_id         = "ru-central1"
}

# Local variable for YDB connection string
locals {
  ydb_connection_string = "grpcs://ydb.serverless.yandexcloud.net:2135/?database=/${var.folder_id}/rsgb-ydb"
  ydb_database_path     = "${var.folder_id}/rsgb-ydb"
}

# Note: Tables are created programmatically by the application on startup
# See common/pkg/ydb/schema.go for table definitions
