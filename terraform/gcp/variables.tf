variable "render_dir" {
  description = "Path to output generated Kubernetes manifests"
  type        = "string"
  default     = ""
}

variable "config_path" {
  description = "Path to ConfigMap describing configuration for config-api"
  type        = "string"
  default     = ""
}

variable "manifests_dir" {
  description = "Directory containing manifests used for deployment"
  type        = "string"
  default     = ""
}

variable "kubeconfig" {
  description = "Path to kubeconfig used to authenticate with Kubernetes API server"
  type        = "string"
}

variable "gcloud_creds" {
  description = "Credentials used to authenticate with Google Cloud and create a cluster. Credentials can be downloaded at https://console.cloud.google.com/apis/credentials/serviceaccountkey."
  type        = "string"
  default     = ""
}

variable "cluster_project" {
  description = "Gcloud project to deploy cluster in."
  type        = "string"
}

variable "cluster_zone" {
  description = "Gcloud zone to deploy in."
  type        = "string"
  default     = "us-west1"
}

variable "sql_service_account_email" {
  description = "Email of the Google Service Account used to access SQL database instances."
  type        = "string"
}

variable "sql_connection_name" {
  description = "Hostname provided by Google to connect to a Cloud SQL instance"
  type        = "string"
}

variable "sql_instance_id" {
  description = "ID of the Cloud SQL instance"
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
