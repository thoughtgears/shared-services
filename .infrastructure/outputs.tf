output "service_account_emails" {
  value       = local.service_account_email_map
  description = "A map of service account emails for the services, where the key is the service name and the value is the service account email."
}

output "document_bucket" {
  value       = google_storage_bucket.run_documents.name
  description = "The name of the document bucket."
}

output "api_gateway_url" {
  description = "Default hostname URL of the deployed API Gateway"
  value       = "https://${google_api_gateway_gateway.gateway.default_hostname}"
}

output "api_config_name" {
  description = "Name of the deployed API Config"
  value       = google_api_gateway_api_config.api_config.name
}