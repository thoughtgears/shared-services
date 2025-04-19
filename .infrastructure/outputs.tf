output "service_account_email" {
  value       = google_service_account.run.email
  description = "A map of service account emails for the services, where the key is the service name and the value is the service account email."
}

output "document_bucket" {
  value       = google_storage_bucket.run_documents.name
  description = "The name of the document bucket."
}
