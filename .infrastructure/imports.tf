import {
  id = "${var.project_id}/${var.project_id}-documents"
  to = google_storage_bucket.run_documents
}

import {
  id = "projects/${var.project_id}/locations/${var.region}/repositories/apis"
  to = google_artifact_registry_repository.apis
}

import {
  id = "projects/${var.project_id}/locations/${var.region}/repositories/utils"
  to = google_artifact_registry_repository.utils
}