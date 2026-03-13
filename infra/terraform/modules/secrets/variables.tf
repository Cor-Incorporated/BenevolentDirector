variable "secret_names" {
  description = "Names of Secret Manager secrets to create"
  type        = list(string)

  validation {
    condition     = length(var.secret_names) == length(toset(var.secret_names))
    error_message = "secret_names must not contain duplicates."
  }
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}
