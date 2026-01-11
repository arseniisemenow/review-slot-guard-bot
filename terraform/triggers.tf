# Trigger for periodic job execution
resource "yandex_function_trigger" "periodic_job" {
  folder_id  = var.folder_id
  name       = "rsgb-periodic-trigger"
  description = "Trigger PeriodicJob every 5 minutes"

  timer {
    # Format: Minutes Hours Day-of-month Month Day-of-week Year
    # Use ? in Day-of-week when Day-of-month is specified (cannot use both *)
    # Examples from docs: "* * * * ? *" = every minute
    cron_expression = "*/5 * * * ? *"
  }

  function {
    id                = yandex_function.periodic_job.id
    service_account_id = yandex_iam_service_account.review_slot_guard_bot.id
  }
}
