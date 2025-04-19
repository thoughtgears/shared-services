output "service_account_email" {
  value       = google_service_account.run.email
  description = "A map of service account emails for the services, where the key is the service name and the value is the service account email."
}

output "document_bucket" {
  value       = google_storage_bucket.run_documents.name
  description = "The name of the document bucket."
}

output "domain_mapping_records" {
  value       = try(google_cloud_run_domain_mapping.this.status[0].resource_records, [])
  description = "The domain mapping records for the Cloud Run service."
}