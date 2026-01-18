# Service Account for the bot
resource "yandex_iam_service_account" "review_slot_guard_bot" {
  folder_id = var.folder_id
  name      = "rsgb-sa"
}

# IAM Role Assignments
resource "yandex_resourcemanager_folder_iam_member" "ydb_editor" {
  folder_id   = var.folder_id
  role        = "ydb.editor"
  member      = "serviceAccount:${yandex_iam_service_account.review_slot_guard_bot.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "functions_invoker" {
  folder_id   = var.folder_id
  role        = "serverless.functions.invoker"
  member      = "serviceAccount:${yandex_iam_service_account.review_slot_guard_bot.id}"
}
