terraform {
  required_providers {
    aws = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }

  required_version = ">= 1.10.0"
}
