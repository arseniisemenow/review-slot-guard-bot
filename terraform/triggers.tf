# Trigger for periodic job execution
resource "yandex_function_trigger" "periodic_job" {
  folder_id  = var.folder_id
  name       = "rsgb-periodic-trigger"
  description = "Trigger PeriodicJob every 5 minutes"

  timer {
    cron_expression = var.periodic_job_schedule
  }

  function {
    id = yandex_function.periodic_job.id
  }
}
