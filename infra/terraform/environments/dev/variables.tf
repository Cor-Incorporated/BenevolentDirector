variable "project_id" {
  description = "GCP project ID for the dev environment"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-northeast1"
}

variable "developer_cidr_blocks" {
  description = "Developer CIDR blocks for GKE master access (e.g. [\"203.0.113.5/32\"]). Pass via TF_VAR_developer_cidr_blocks to avoid committing IPs."
  type        = list(string)
  default     = []

  validation {
    condition     = alltrue([for cidr in var.developer_cidr_blocks : can(cidrhost(cidr, 0))])
    error_message = "Each entry must be a valid CIDR block (e.g. \"203.0.113.5/32\")."
  }
}
