resource "spectrocloud_virtual_machine" "sivaTestVM" {
  // noproxy-san-frascico
#  cluster_uid = "64063a05f47704b6e2a856dc"
  // shruthi dev env
  cluster_uid = "6406b8ee7c865766fde23247"
  name = "siva-terraform-read"
  namespace = "default"
  cpu_cores = 2
  run_on_launch = true
  memory = "4G"
  // image_url is has conflict with volume
  image_url = "quay.io/kubevirt/fedora-cloud-container-disk-demo"

  // For now we have gave support only for container_disk/cloud_init_no_cloud
#  devices {
#    disk {
#      name = "containerdisk"
#      bus  = "virtio"
#    }
#    disk {
#      name = "cloudinitdisk"
#      bus  = "virtio"
#    }
#    interface {
#      name = "default"
#    }
#  }
#
#  volume {
#    name = "containerdisk"
#    container_disk {
#      image_url = "quay.io/kubevirt/fedora-cloud-container-disk-demo"
#    }
#  }
#  volume {
#    name = "cloudinitdisk"
#    cloud_init_no_cloud {
#      user_data = "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n"
#    }
#  }
#
  network {
    name = "default"
  }
}