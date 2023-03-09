resource "spectrocloud_virtual_machine" "tf-test-vm-default" {
  cluster_uid = "6406b8ee7c865766fde23247"
  name = "tf-test-vm-default"
  namespace = "default"
  cpu_cores = 2
  memory = "3G"
  image_url = "quay.io/kubevirt/fedora-cloud-container-disk-demo"
  vm_action = "migrate"
  labels = ["tf-test=true", "env=dev"]
  annotations = {
    tf : "true",
    owner: "siva",
  }
}

resource "spectrocloud_virtual_machine" "tf-test-vm-custom" {
  cluster_uid = "6406b8ee7c865766fde23247"
  name = "tf-test-vm-custom"
  namespace = "default"
  cpu_cores = 2
  memory = "3G"
  // For now we have gave support only for container_disk/cloud_init_no_cloud
  devices {
    disk {
      name = "containerdisk1"
      bus  = "virtio"
    }
    disk {
      name = "cloudinitdisk1"
      bus  = "virtio"
    }
    interface {
      name = "default"
    }
  }
  volume {
    name = "containerdisk1"
    container_disk {
      image_url = "quay.io/kubevirt/fedora-cloud-container-disk-demo"
    }
  }
  volume {
    name = "cloudinitdisk1"
    cloud_init_no_cloud {
      user_data = "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n"
    }
  }
  network {
    name = "default"
  }
}
