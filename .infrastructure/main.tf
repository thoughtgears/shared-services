locals {
  service_name            = "portal-api"
  otel_docker_image       = "${var.region}-docker.pkg.dev/${var.project_id}/utils/otel:latest"
  portal_api_docker_image = "${var.region}-docker.pkg.dev/${var.project_id}/apis/${local.service_name}@${var.digest}"

  service_apis = [
    "apigateway.googleapis.com",
    "servicemanagement.googleapis.com",
    "servicecontrol.googleapis.com",
    "run.googleapis.com",
    "artifactregistry.googleapis.com"
  ]
}

/**
 * # Initial configuration
 *
 * This section enables the necessary APIs for the project.
 * It uses a local variable to define the APIs to be enabled.
 */
resource "google_project_service" "this" {
  for_each = toset(local.service_apis)

  project = var.project_id
  service = each.value

  disable_dependent_services = false
  disable_on_destroy         = false
}

/**
 * # Cloud run resources
 *
 * This section sets up all infrastructure needed for the cloud run instances.
 * It creates a service account for each service, and assigns the necessary IAM roles.
 * It also creates a custom IAM role for Firebase service access.
 */
resource "google_project_iam_custom_role" "thoughtgears_firebase_service_access" {
  project     = var.project_id
  role_id     = "thougthgears.documentDatastoreAccess"
  title       = "Document Datastore Access"
  description = "Thoughtgears role for document API to access Datastore/Firestore"
  permissions = [
    "datastore.entities.get",
    "datastore.entities.list",
    "datastore.entities.create",
    "datastore.entities.update",
    "datastore.entities.delete",
    "datastore.databases.get"
  ]
}

resource "google_service_account" "run" {
  project      = var.project_id
  account_id   = "run-${local.service_name}"
  display_name = "[RUN] ${local.service_name}"
  description  = "Service account for cloud run instance for ${local.service_name}"
}

resource "google_secret_manager_secret" "firebase_json_key" {
  project   = var.project_id
  secret_id = "firebase-json-key"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "firebase_json_key" {
  secret = google_secret_manager_secret.firebase_json_key.id

  is_secret_data_base64 = true
  secret_data           = filebase64("../secrets/firebase-service-account.json")
}

resource "google_cloud_run_v2_service" "this" {
  name                = local.service_name
  project             = var.project_id
  location            = var.region
  deletion_protection = false
  ingress             = "INGRESS_TRAFFIC_ALL"

  template {
    service_account                  = google_service_account.run.email
    timeout                          = "60s"
    max_instance_request_concurrency = 500

    volumes {
      name = "secret-volume"
      secret {
        secret       = google_secret_manager_secret.firebase_json_key.secret_id
        default_mode = 292 # 0444
        items {
          version = "latest"
          path    = "firebase-service-account.json"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    containers {
      image = local.portal_api_docker_image

      volume_mounts {
        name       = "secret-volume"
        mount_path = "/secrets"
      }

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }

      env {
        name  = "GCP_BUCKET_NAME"
        value = google_storage_bucket.run_documents.name
      }

      env {
        name  = "GCP_PROJECT_ID"
        value = var.project_id
      }

      env {
        name  = "GCP_REGION"
        value = var.region
      }

      env {
        name  = "DOMAIN_NAME"
        value = "thoughtgears.dev"
      }

      env {
        name  = "LOCAL"
        value = false
      }

      env {
        name  = "OTEL_ENDPOINT"
        value = "localhost:4317"
      }

      env {
        name  = "GIN_MODE"
        value = "release"
      }
    }

    containers {
      name  = "otel"
      image = local.otel_docker_image

      resources {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }
    }
  }
}

/**
 * # IAM roles for service accounts
 *
 * This section assigns the necessary IAM roles to the service accounts.
 */
resource "google_project_iam_binding" "run_firebase_access" {
  project = var.project_id
  members = ["serviceAccount:${google_service_account.run.email}"]
  role    = google_project_iam_custom_role.thoughtgears_firebase_service_access.name
}

resource "google_storage_bucket_iam_binding" "run_object_admin" {
  bucket  = google_storage_bucket.run_documents.name
  members = ["serviceAccount:${google_service_account.run.email}"]
  role    = "roles/storage.objectAdmin"
}

resource "google_secret_manager_secret_iam_binding" "run_firebase_json_key" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.firebase_json_key.secret_id
  members   = ["serviceAccount:${google_service_account.run.email}"]
  role      = "roles/secretmanager.secretAccessor"
}

resource "google_project_iam_member" "run_firebase_admin" {
  project = var.project_id
  member  = "serviceAccount:${google_service_account.run.email}"
  role    = "roles/firebase.admin"
}

resource "google_project_iam_member" "run_trace_writer" {
  project = var.project_id
  member  = "serviceAccount:${google_service_account.run.email}"
  role    = "roles/telemetry.writer"
}

resource "google_project_iam_member" "run_trace_agent" {
  project = var.project_id
  member  = "serviceAccount:${google_service_account.run.email}"
  role    = "roles/cloudtrace.agent"
}

resource "google_project_iam_member" "run_service_account_user" {
  project = var.project_id
  member  = "serviceAccount:${google_service_account.run.email}"
  role    = "roles/iam.serviceAccountUser"
}

resource "google_project_iam_member" "run_service_usage_consumer" {
  project = var.project_id
  member  = "serviceAccount:${google_service_account.run.email}"
  role    = "roles/serviceusage.serviceUsageConsumer"
}

resource "google_project_iam_member" "run_metrics_writer" {
  project = var.project_id
  member  = "serviceAccount:${google_service_account.run.email}"
  role    = "roles/monitoring.metricWriter"
}

resource "google_project_iam_member" "run_artifact_registry_reader" {
  project = var.project_id
  member  = "serviceAccount:${google_service_account.run.email}"
  role    = "roles/artifactregistry.reader"
}


resource "google_cloud_run_v2_service_iam_member" "monitoring_uptime_check_invoker" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.this.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:service-${var.project_number}@gcp-sa-monitoring-notification.iam.gserviceaccount.com"
}

resource "google_cloud_run_v2_service_iam_member" "run_allow_all_users" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.this.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}


/**
 * # GCS infrastructure
 *
 * This section manages buckets for the APIs.
 */
resource "google_storage_bucket" "run_documents" {
  project                     = var.project_id
  location                    = "EU"
  name                        = "${var.project_id}-documents"
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  force_destroy               = false
}

/**
 * # Artifact registry
 *
 * This section sets up the artifact registry for the APIs.
 */

resource "google_artifact_registry_repository" "apis" {
  project  = var.project_id
  location = var.region

  repository_id          = "apis"
  format                 = "DOCKER"
  description            = "Artifact registry for APIs"
  cleanup_policy_dry_run = true

  docker_config {
    immutable_tags = false
  }

  depends_on = [google_project_service.this["artifactregistry.googleapis.com"]]
}

resource "google_artifact_registry_repository" "utils" {
  project  = var.project_id
  location = var.region

  repository_id          = "utils"
  format                 = "DOCKER"
  description            = "Artifact registry for Utils"
  cleanup_policy_dry_run = true

  docker_config {
    immutable_tags = false
  }

  depends_on = [google_project_service.this["artifactregistry.googleapis.com"]]
}
