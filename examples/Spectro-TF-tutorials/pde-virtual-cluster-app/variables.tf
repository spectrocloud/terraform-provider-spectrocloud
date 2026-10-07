variable "cluster-group-name" {
  type        = string
  description = "The name of an existing cluster group (created via Palette Dev Engine / App Mode) to host the virtual cluster(s)."
}

variable "scenario-one-cluster-name" {
  type        = string
  default     = "cluster-1"
  description = "The name of the virtual cluster used for scenario 1 (single-container UI app)."
}

variable "scenario-two-cluster-name" {
  type        = string
  default     = "cluster-2"
  description = "The name of the virtual cluster used for scenario 2 (three-tier app), when enable-second-scenario is true."
}

variable "single-container-image" {
  type        = string
  description = "The container image for the standalone Hello Universe UI app in scenario 1."
  default     = "ghcr.io/spectrocloud/hello-universe:1.1.3"
}

variable "multiple_container_images" {
  type        = map(string)
  description = "The UI and API container images for the three-tier Hello Universe app in scenario 2."
  default = {
    ui  = "ghcr.io/spectrocloud/hello-universe:1.1.3"
    api = "ghcr.io/spectrocloud/hello-universe-api:1.0.12"
  }
}

variable "database-version" {
  type        = string
  description = "The Postgres version for scenario 2's database tier."
  default     = "14"
}

variable "database-name" {
  type        = string
  description = "The database name for scenario 2's database tier."
  default     = "counter"
}

variable "database-user" {
  type        = string
  description = "The database user for scenario 2's database tier."
  default     = "pguser"
}

variable "database-ssl-mode" {
  type        = string
  description = "The SSL mode used when the scenario 2 API tier connects to the database."
  default     = "require"
}

variable "token" {
  type        = string
  default     = "931A3B02-8DCC-543F-A1B2-69423D1A0B94"
  description = "A fixed anonymous usage token baked into the Hello Universe UI app (scenario 2) — shared across all deployments of this tutorial, not a per-user credential."
}

variable "enable-second-scenario" {
  type        = bool
  description = "Whether to also deploy scenario 2 (the three-tier app: UI + API + Postgres) alongside scenario 1."
  default     = false
}

variable "tags" {
  type        = list(string)
  description = "The default tags to apply to Palette resources."
  default = [
    "spectro-cloud-education",
    "app:hello-universe",
    "repository:spectrocloud/tutorials/",
    "terraform_managed:true",
    "tutorial:hello-universe-tf"
  ]
}
