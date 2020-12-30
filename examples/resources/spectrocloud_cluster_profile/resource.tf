resource "spectrocloud_cluster_profile" "cp-addon-azure" {
  name        = "cp-basic"
  description = "basic cp"
  cloud       = "azure"
  type        = "add-on"

  pack {
    name = "spectro-byo-manifest"
    tag  = "1.0.x"
    uid  = "5faad584f244cfe0b98cf489"
    # layer  = ""
    values = <<-EOT
      manifests:
        byo-manifest:
          contents: |
            # Add manifests here
            apiVersion: v1
            kind: Namespace
            metadata:
              labels:
                app: wordpress
                app3: wordpress3
              name: wordpress
    EOT
  }

}
