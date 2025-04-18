variable "project_id" {
  type        = string
  description = "GCP Project ID to deploy the infrastructure"
}

variable "project_number" {
  type        = string
  description = "GCP Project Number (needed for API Gateway Service Account IAM)"
}

variable "region" {
  type        = string
  description = "GCP Region to deploy the infrastructure"
  default     = "europe-west1"
}
