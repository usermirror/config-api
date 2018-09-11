variable "render_dir" {
  description = "Path to directory where templated manifests can be outputted"
  type        = "string"
  default     = "../dist/manifests"
}

variable "config_path" {
  description = "Path to ConfigMap describing configuration for config-api"
  type        = "string"
  default     = "../manifests/config.yaml"
}

variable "manifests_dir" {
  description = "Directory containing manifests used for deployment"
  type        = "string"
  default     = "../manifests"
}

variable "kubeconfig" {
  description = "Path to kubeconfig used to authenticate with Kubernetes API server"
  type        = "string"
}
