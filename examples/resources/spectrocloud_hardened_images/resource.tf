# Hold two clusters back from the hardened image rollout.
# Every other cluster in the tenant receives the Spectro Cloud registry image pull secret.
resource "spectrocloud_hardened_images" "hardened_images" {
  excluded_cluster_uids = [
    "6512f1a3b0e4d8c7a9f21b40",
    "6512f1a3b0e4d8c7a9f21b57",
  ]
}

# Surface the clusters Palette could not deliver the image pull secret to
output "failed_hardened_image_clusters" {
  value = [
    for cluster in spectrocloud_hardened_images.hardened_images.clusters :
    cluster.name if cluster.state == "Failed"
  ]
}

## import the existing tenant hardened image configuration
#import {
#  to = spectrocloud_hardened_images.hardened_images
#  id = "{tenantUID}" // tenant-uid or organization name
#}
