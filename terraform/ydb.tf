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

# Table: users
resource "yandex_ydb_table" "users" {
  connection_string = local.ydb_connection_string
  path              = "${local.ydb_database_path}/users"
  depends_on        = [yandex_ydb_database_serverless.review_slot_guard_bot]

  column {
    name     = "reviewer_login"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "status"
    type     = "Utf8"
    not_null = true
  }
  column {
    name = "telegram_chat_id"
    type = "Int64"
  }
  column {
    name     = "created_at"
    type     = "Datetime"
    not_null = true
  }
  column {
    name     = "last_auth_success_at"
    type     = "Datetime"
    not_null = true
  }
  column {
    name = "last_auth_failure_at"
    type = "Datetime"
  }

  primary_key = ["reviewer_login"]
}

# Table: user_settings
resource "yandex_ydb_table" "user_settings" {
  connection_string = local.ydb_connection_string
  path              = "${local.ydb_database_path}/user_settings"
  depends_on        = [yandex_ydb_database_serverless.review_slot_guard_bot]

  column {
    name     = "reviewer_login"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "response_deadline_shift_minutes"
    type     = "Int32"
    not_null = true
  }
  column {
    name     = "non_whitelist_cancel_delay_minutes"
    type     = "Int32"
    not_null = true
  }
  column {
    name     = "notify_whitelist_timeout"
    type     = "Bool"
    not_null = true
  }
  column {
    name     = "notify_non_whitelist_cancel"
    type     = "Bool"
    not_null = true
  }
  column {
    name     = "slot_shift_threshold_minutes"
    type     = "Int32"
    not_null = true
  }
  column {
    name     = "slot_shift_duration_minutes"
    type     = "Int32"
    not_null = true
  }
  column {
    name     = "cleanup_durations_minutes"
    type     = "Int32"
    not_null = true
  }

  primary_key = ["reviewer_login"]
}

# Table: user_project_whitelist
resource "yandex_ydb_table" "user_project_whitelist" {
  connection_string = local.ydb_connection_string
  path              = "${local.ydb_database_path}/user_project_whitelist"
  depends_on        = [yandex_ydb_database_serverless.review_slot_guard_bot]

  column {
    name     = "reviewer_login"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "entry_type"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "name"
    type     = "Utf8"
    not_null = true
  }

  primary_key = ["reviewer_login", "entry_type", "name"]
}

# Table: project_families
resource "yandex_ydb_table" "project_families" {
  connection_string = local.ydb_connection_string
  path              = "${local.ydb_database_path}/project_families"
  depends_on        = [yandex_ydb_database_serverless.review_slot_guard_bot]

  column {
    name     = "family_label"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "project_name"
    type     = "Utf8"
    not_null = true
  }

  primary_key = ["family_label", "project_name"]
}

# Table: review_requests
resource "yandex_ydb_table" "review_requests" {
  connection_string = local.ydb_connection_string
  path              = "${local.ydb_database_path}/review_requests"
  depends_on        = [yandex_ydb_database_serverless.review_slot_guard_bot]

  column {
    name     = "id"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "reviewer_login"
    type     = "Utf8"
    not_null = true
  }
  column {
    name = "notification_id"
    type = "Utf8"
  }
  column {
    name = "project_name"
    type = "Utf8"
  }
  column {
    name = "family_label"
    type = "Utf8"
  }
  column {
    name     = "review_start_time"
    type     = "Datetime"
    not_null = true
  }
  column {
    name     = "calendar_slot_id"
    type     = "Utf8"
    not_null = true
  }
  column {
    name = "decision_deadline"
    type = "Datetime"
  }
  column {
    name = "non_whitelist_cancel_at"
    type = "Datetime"
  }
  column {
    name = "telegram_message_id"
    type = "Utf8"
  }
  column {
    name     = "status"
    type     = "Utf8"
    not_null = true
  }
  column {
    name     = "created_at"
    type     = "Datetime"
    not_null = true
  }
  column {
    name = "decided_at"
    type = "Datetime"
  }

  primary_key = ["id"]
}
