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
  description = "Developer IP CIDR blocks for GKE master access (pass via .tfvars or TF_VAR_)"
  type        = list(string)
  default     = []
}
