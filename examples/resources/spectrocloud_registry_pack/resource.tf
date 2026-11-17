# Pack registry with basic authentication.
resource "spectrocloud_registry_pack" "pack_registry" {
  name     = "my-pack-registry"
  endpoint = "https://registry.example.com:5000"

  credentials {
    credential_type = "basic"
    username        = var.pack_registry_username
    password        = var.pack_registry_password

    # Optional: TLS configuration. Omit the block to send no TLS configuration.
    # Certificates are PEM content, not paths - use file() to read them from disk.
    # tls_config {
    #   ca                   = file("${path.module}/ca.pem")
    #   insecure_skip_verify = false
    # }
  }

  # Optional: wait until Palette has synchronized the registry's packs.
  wait_for_sync = true
}

# Pack registry that needs no authentication.
resource "spectrocloud_registry_pack" "public_pack_registry" {
  name     = "public-pack-registry"
  endpoint = "https://public-registry.example.com"

  credentials {
    credential_type = "noAuth"
  }
}

variable "pack_registry_username" {
  type = string
}

variable "pack_registry_password" {
  type      = string
  sensitive = true
}
