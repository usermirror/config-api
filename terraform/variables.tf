variable "render_dir" {
  description = "Path to directory where templated manifests can be outputted (defaults to path within module)"
  type        = "string"
  default     = ""
}

variable "config_path" {
  description = "Path to ConfigMap describing configuration for config-api (defaults to path within module)"
  type        = "string"
  default     = ""
}

variable "manifests_dir" {
  description = "Directory containing manifests used for deployment (defaults to path within module)"
  type        = "string"
  default     = ""
}

variable "kubeconfig" {
  description = "Path to kubeconfig used to authenticate with Kubernetes API server"
  type        = "string"
}

variable "sql_db_password" {
  description = "Password for the SQL user used to perform writes."
  type        = "string"
  default     = "changeme"
}

variable "image_repo" {
  description = "Image repository for config-api"
  type        = "string"
  default     = "us.gcr.io/usermirror-staging/config-api"
}

variable "image_tag" {
  description = "Image tag for config-api"
  type        = "string"
  default     = "latest"
}

variable "domain" {
  description = "Domain which is used in configuring ingresses"
  type        = "string"
  default     = "staging.helpusersvote.com"
}

variable "tls_secret" {
  description = "Name of Secret containing TLS credentials used in configuring ingresses"
  type        = "string"
  default     = "ingress-tls"
}

variables "replicas" {
  description = "Number of config-api Pods to be deployed"
  type        = "string"
  default     = "3"
}

variable "last_resource" {
  description = "Allows dependency to be expressed to module"
  type        = "string"
  default     = ""
}
