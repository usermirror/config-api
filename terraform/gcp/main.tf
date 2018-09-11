provider "google" {
  version = "~> 1.17"

  credentials = "${var.gcloud_creds}"
  project     = "${var.cluster_project}"
  region      = "${var.cluster_zone}"
}

// defaults requiring interpolation
locals {
  render_dir    = "${var.render_dir=="" ? local.default_render_dir : var.render_dir}"
  config_path   = "${var.config_path=="" ? local.default_config_path : var.config_path}"
  manifests_dir = "${var.manifests_dir=="" ? local.default_manifests_dir : var.manifests_dir}"

  default_render_dir    = "${dirname(dirname(path.module))}/dist/manifests"
  default_config_path   = "${dirname(dirname(path.module))}/manifests/config.yaml"
  default_manifests_dir = "${dirname(dirname(path.module))}/manifests"
}

module "config-api" {
  source = "../"

  kubeconfig    = "${var.kubeconfig}"
  config_path   = "${local.config_path}"
  manifests_dir = "${local.manifests_dir}"
  render_dir    = "${local.render_dir}"

  last_resource = "${module.cloudsql_db.sql_access_key}"
}

// cloudsql_db creates a user, database, and access credentials on a PostgreSQL instance.
module "cloudsql_db" {
  source = "git::https://github.com/helpusersvote/terraform-kubernetes-helpusersvote.git//modules/cloudsql_db"

  render_dir = "${local.manifests_dir}/config-api"

  client_service_account_email = "${var.sql_service_account_email}"
  connection_name              = "${var.sql_connection_name}"
  instance                     = "${var.sql_instance_id}"

  db_user          = "huv_user"
  db_user_password = "${var.sql_db_password}"
  db_name          = "huv_db"
}
