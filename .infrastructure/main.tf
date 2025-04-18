locals {
  services = {
    "user-api" = {
      name        = "user-api"
      description = "User API"
    }
    "document-api" = {
      name        = "document-api"
      description = "Document API"
    }
  }

  otel_docker_image = "europe-west1-docker.pkg.dev/shared-services-47252/utils/otel:latest"

  service_account_list = [
    for key, service in local.services :
    "serviceAccount:run-${service.name}@${var.project_id}.iam.gserviceaccount.com"
  ]

  service_account_email_map = {
    for key, service in local.services :
    service.name => "run-${service.name}@${var.project_id}.iam.gserviceaccount.com"
  }

  service_apis = [
    "apigateway.googleapis.com",
    "servicemanagement.googleapis.com",
    "servicecontrol.googleapis.com",
    "run.googleapis.com",
    "artifactregistry.googleapis.com"
  ]
}

locals {
  api_gateway_runtime_sa = "service-${var.project_number}@gcp-sa-apigateway.iam.gserviceaccount.com"
  openapi_spec_rendered = templatefile("${path.module}/templates/openapi.yaml.tftpl", {
    project_id         = var.project_id
    cloud_run_user_url = google_cloud_run_v2_service.this["user-api"].uri
    cloud_run_doc_url  = google_cloud_run_v2_service.this["document-api"].uri
  })

  openapi_spec_base64 = base64encode(local.openapi_spec_rendered)
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
  for_each = local.services

  project      = var.project_id
  account_id   = "run-${each.value.name}"
  display_name = "[RUN] ${each.value.name}"
  description  = "Service account for cloud run instance for ${each.value.description}"
}

resource "google_cloud_run_v2_service" "this" {
  for_each = local.services

  name                = each.value.name
  project             = var.project_id
  location            = var.region
  deletion_protection = false
  ingress             = "INGRESS_TRAFFIC_ALL"

  template {
    service_account                  = google_service_account.run[each.key].email
    timeout                          = "60s"
    max_instance_request_concurrency = 500

    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/apis/${each.key}:latest"

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      dynamic "env" {
        for_each = var.service_env_vars[each.value.name].env_vars

        content {
          name  = env.key
          value = env.value
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
        name  = "OTEL_ENDPOINT"
        value = "localhost:4317"
      }
    }

    containers {
      name  = "otel"
      image = local.otel_docker_image

      resources {
        limits = {
          cpu    = "1"
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
  members = local.service_account_list
  role    = google_project_iam_custom_role.thoughtgears_firebase_service_access.name
}

resource "google_storage_bucket_iam_binding" "run_object_admin" {
  bucket  = google_storage_bucket.run_documents.name
  members = local.service_account_list
  role    = "roles/storage.objectAdmin"
}

resource "google_project_iam_member" "run_service_account_user" {
  for_each = toset(local.service_account_list)

  project = var.project_id
  member  = each.value
  role    = "roles/iam.serviceAccountUser"
}

resource "google_project_iam_member" "run_service_usage_consumer" {
  for_each = toset(local.service_account_list)

  project = var.project_id
  member  = each.value
  role    = "roles/serviceusage.serviceUsageConsumer"
}

resource "google_project_iam_member" "run_metrics_writer" {
  for_each = toset(local.service_account_list)

  project = var.project_id
  member  = each.value
  role    = "roles/monitoring.metricWriter"
}

resource "google_cloud_run_v2_service_iam_member" "user_api_invoker" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.this["user-api"].name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${local.api_gateway_runtime_sa}"
}

# Grant API Gateway permission to invoke the Document API Cloud Run service
resource "google_cloud_run_v2_service_iam_member" "document_api_invoker" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.this["document-api"].name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${local.api_gateway_runtime_sa}"
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
 * # API Gateway for the run instances
 *
 * This section sets up the API Gateway for the cloud run instances.
 * It creates a global gateway and instance.
 * It also creates a unified gateway configuration for the services.
 */
resource "google_api_gateway_api" "portal_api" {
  provider     = google-beta
  project      = var.project_id
  api_id       = "portal-api"
  display_name = "Portal API"

  depends_on = [
    google_project_service.this["apigateway.googleapis.com"],
    google_project_service.this["servicemanagement.googleapis.com"],
    google_project_service.this["servicecontrol.googleapis.com"],
  ]
}

resource "google_api_gateway_api_config" "api_config" {
  provider             = google-beta
  project              = var.project_id
  api                  = google_api_gateway_api.portal_api.api_id
  api_config_id_prefix = "portal-api-config-" # Creates unique IDs like my-gateway-config-a1b2

  display_name = "Config ${timestamp()}" # Example: Config with timestamp

  openapi_documents {
    document {
      path     = "openapi_spec.yaml" # Arbitrary filename for the spec within the config
      contents = local.openapi_spec_base64
    }
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    google_project_service.this["apigateway.googleapis.com"],
    google_project_service.this["servicemanagement.googleapis.com"],
    google_project_service.this["servicecontrol.googleapis.com"],
    google_cloud_run_v2_service.this["user-api"],
    google_cloud_run_v2_service.this["document-api"],
  ]
}

resource "google_api_gateway_gateway" "gateway" {
  provider   = google-beta
  project    = var.project_id
  region     = var.region
  gateway_id = "portal-api"

  api_config   = google_api_gateway_api_config.api_config.id
  display_name = "Portal API Gateway Instance"

  depends_on = [
    google_api_gateway_api_config.api_config
  ]
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
