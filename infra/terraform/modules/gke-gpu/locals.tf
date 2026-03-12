locals {
  cluster_name   = "bd-${var.environment}-gke-gpu"
  node_pool_name = "bd-${var.environment}-gpu-pool"

  default_labels = {
    environment = var.environment
    project     = "benevolent-director"
    managed_by  = "terraform"
    component   = "gke-gpu"
  }

  labels = merge(local.default_labels, var.labels)

  # Secondary ranges for pods and services
  pod_range_name     = "bd-${var.environment}-gke-pods"
  service_range_name = "bd-${var.environment}-gke-services"
}
