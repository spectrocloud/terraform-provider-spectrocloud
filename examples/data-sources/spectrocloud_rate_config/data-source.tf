data "spectrocloud_rate_config" "current" {}

output "vsphere_cpu_unit_price_per_hour" {
  value = data.spectrocloud_rate_config.current.vsphere[0].cpu_unit_price_per_hour
}

output "aws_compute_optimized_compute_rate_proportion" {
  value = data.spectrocloud_rate_config.current.aws[0].compute_optimized[0].compute_rate_proportion
}
