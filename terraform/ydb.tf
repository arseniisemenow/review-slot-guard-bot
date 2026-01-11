# YDB Serverless Database
resource "yandex_ydb_database_serverless" "review_slot_guard_bot" {
  folder_id           = var.folder_id
  name                = "rsgb-ydb"
  location_id         = "ru-central1"
}

# Tables will be created programmatically by the application
# The YDB tables API in terraform has changed, so we create the database
# and let the application handle table creation via SQL
