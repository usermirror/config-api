// defaults requiring interpolation
locals {
  render_dir    = "${var.render_dir=="" ? local.default_render_dir : var.render_dir}"
  config_path   = "${var.config_path=="" ? local.default_config_path : var.config_path}"
  manifests_dir = "${var.manifests_dir=="" ? local.default_manifests_dir : var.manifests_dir}"

  default_render_dir    = "${dirname(path.module)}/dist/manifests"
  default_config_path   = "${dirname(path.module)}/manifests/config.yaml"
  default_manifests_dir = "${dirname(path.module)}/manifests"

  last = "${element(list("", var.last_resource), 0)}"
}

module "config" {
  source = "git::https://github.com/helpusersvote/terraform-kubernetes-helpusersvote.git//modules/config?ref=85dac1099a4820c9c78be90956bbfc35f7f9df2f"

  components   = ["embed-config-api", "ballot-api"]
  render_dir   = "${local.render_dir}"
  config       = "${local.config_path}${local.last}" // TODO: remove once module dependency can be improved
  manifest_dir = "${local.manifests_dir}"
  git_dir      = "${path.module}"

  vars = {
    sql_db_password = "${var.sql_db_password}"

    image_repo = "${var.image_repo}"

    domain     = "${var.domain}"
    tls_secret = "${var.tls_secret}"
  }
}

module "kubernetes" {
  source = "git::https://github.com/helpusersvote/terraform-kubernetes-helpusersvote.git//modules/kubernetes?ref=85dac1099a4820c9c78be90956bbfc35f7f9df2f"

  manifest_dirs = "${module.config.dirs}"
  kubeconfig    = "${var.kubeconfig}"
  last_resource = "${join(",", module.config.manifest_state)}"
}
