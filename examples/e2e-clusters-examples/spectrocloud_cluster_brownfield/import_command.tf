# Writes the generated kubectl import command to local disk on every apply, so it's easy to find
# and run against the real, already-existing cluster to complete the import. Uses the
# hashicorp/local provider - the same mechanism the other end-to-end examples use for
# kubeconfig.tf, adapted here since brownfield clusters expose no kubeconfig of their own.
#
# Day-2 note: spectrocloud_cluster_brownfield.cluster.kubectl_command is Computed - Palette
# regenerates it whenever the registration changes, so this file is rewritten on every apply.
resource "local_file" "import_command" {
  content         = spectrocloud_cluster_brownfield.cluster.kubectl_command
  filename        = "${var.import_command_output_dir}/import_${spectrocloud_cluster_brownfield.cluster.name}.sh"
  file_permission = "0700"
}
