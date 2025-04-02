# Retrieve details of a specific pack using name and version
data "spectrocloud_pack" "example" {
  name    = "nginx-pack" # Pack name (e.g., "nginx-pack", "k8s-core", "monitoring-stack")
  version = "1.2.3"      # Pack version (e.g., "1.2.3", "latest", "stable")
}

# Retrieve a pack using advanced filters
data "spectrocloud_pack" "filtered" {
  name = "k8sgpt-operator" # Pack name to search for

  advance_filters {
    pack_type   = ["spectro"]    # Allowed: "helm", "spectro", "oci", "manifest"
    addon_type  = ["system app"] # Allowed: "load balancer", "ingress", "logging", "monitoring", "security", "authentication", "servicemesh", "system app", "app services", "registry", "csi", "cni", "integration"
    pack_layer  = ["addon"]      # Allowed: "kernel", "os", "k8s", "cni", "csi", "addon"
    environment = ["all"]        # Allowed: "all", "aws", "eks", "azure", "aks", "gcp", "gke", "vsphere", "maas", "openstack", "edge-native"
    is_fips     = false          # Boolean: true (FIPS-compliant) / false (default)
    pack_source = ["community"]  # Allowed: "spectrocloud", "community"
  }

  registry_uid = "5ee9c5adc172449eeb9c30cf" # Unique registry identifier
}

# Output pack details
output "pack_id" {
  value = data.spectrocloud_pack.example.id # Returns the unique pack ID
}

output "pack_version" {
  value = data.spectrocloud_pack.example.version # Returns the pack version
}

output "pack_values" {
  value = data.spectrocloud_pack.example.values # Returns the YAML values of the pack
}