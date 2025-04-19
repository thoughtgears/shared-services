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

variable "git_sha" {
  type        = string
  description = "Git SHA of the commit to deploy"
  default     = "latest"
}

variable "digest" {
  type        = string
  description = "Docker digest of the image to deploy"
  default     = "latest"
}

variable "cors_function_url" {
  type        = string
  description = "URL of the CORS function that handle cors for the API Gateway"
}