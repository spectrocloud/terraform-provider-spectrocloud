# Retrieve details of a specific pack using name and version
#data "spectrocloud_pack" "example" {
#  name    = "nginx-pack" # Pack name (e.g., "nginx-pack", "k8s-core", "monitoring-stack")
#  version = "1.2.3"      # Pack version (e.g., "1.2.3", "latest", "stable")
#}

# Retrieve a pack using advanced filters
data "spectrocloud_pack" "filtered" {
  name = "ubuntu-aws" # Pack name to search for

  advance_filters {
    pack_type   = ["spectro"]    # Allowed: "helm", "spectro", "oci", "manifest"
#    addon_type  = ["system app"] # Allowed: "load balancer", "ingress", "logging", "monitoring", "security", "authentication", "servicemesh", "system app", "app services", "registry", "csi", "cni", "integration"
    pack_layer  = ["os"]      # Allowed: "kernel", "os", "k8s", "cni", "csi", "addon"
    environment = ["aws"]        # Allowed: "all", "aws", "eks", "azure", "aks", "gcp", "gke", "vsphere", "maas", "openstack", "edge-native"
#    is_fips     = false          # Boolean: true (FIPS-compliant) / false (default)
#    pack_source = ["spectrocloud"]  # Allowed: "spectrocloud", "community"
  }

#  registry_uid = "5e2031962f090e2d3d8a3290" # Unique registry identifier
}

# Output pack details
output "pack_id" {
  value = data.spectrocloud_pack.filtered.id # Returns the unique pack ID
}

output "pack_version" {
  value = data.spectrocloud_pack.filtered.version # Returns the pack version
}

output "pack_values" {
  value = data.spectrocloud_pack.filtered.values # Returns the YAML values of the pack
}