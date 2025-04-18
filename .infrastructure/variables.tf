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

variable "service_env_vars" {
  type = map(object({
    env_vars = map(string)
  }))
  description = <<EOT
  Map of environment variables for each service, where the key is the service name and the value is an object containing the environment variables.
      The environment variables are defined as a map of key-value pairs.

      Example:
      service_env_vars = {
          "document-api" = {
          env_vars = {
              "ENV_VAR_1" = "value1"
              "ENV_VAR_2" = "value2"
          }
          }
          "user-api" = {
          env_vars = {
              "ENV_VAR_3" = "value3"
              "ENV_VAR_4" = "value4"
          }
          }
      }
  EOT
  default     = {}
}