data "spectrocloud_cluster" "vm-enabled-cluster" {
  name = "newyork"
  context = "tenant"
}

resource "spectrocloud_virtual_machine" "tf-test-vm-default" {
  cluster_uid = data.spectrocloud_cluster.vm-enabled-cluster.id //"6406b8ee7c865766fde23247"
  name = "tf-test-vm-default"
  namespace = "default"
  cpu_cores = 1
  memory = "2G"
  image_url = "quay.io/kubevirt/fedora-cloud-container-disk-demo"
  labels = ["tf-test=true", "env=sumit-dev"]
  annotations = {
    tf : "true",
    owner: "siva1",
  }
}

resource "spectrocloud_virtual_machine" "tf-test-vm-clone-default" {
  cluster_uid = data.spectrocloud_cluster.vm-enabled-cluster.id
  clone_name = spectrocloud_virtual_machine.tf-test-vm-default.name
  name = "tf-test-vm-clone-default"
  namespace = "default"
  depends_on = [spectrocloud_virtual_machine.tf-test-vm-default]
}

resource "spectrocloud_virtual_machine" "tf-test-vm-custom" {
  cluster_uid = data.spectrocloud_cluster.vm-enabled-cluster.id
  name = "tf-test-vm-custom"
  namespace = "default"
  cpu_cores = 1
  memory = "2G"
  labels = ["tf-test=true", "env=sumit-dev"]
  annotations = {
    tf : "true",
    owner: "siva1",
  }
  devices {
    disk {
      name = "containerdisk0"
      bus  = "virtio"
    }
    disk {
      name = "cloudinitdisk0"
      bus  = "virtio"
    }
    interface {
      name = "default"
    }
  }
  volume_spec{
    volume {
      name = "containerdisk0"
      container_disk {
        image_url = "quay.io/kubevirt/fedora-cloud-container-disk-demo"
      }
    }
    volume {
      name = "cloudinitdisk0"
      cloud_init_no_cloud {
        user_data = "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n"
      }
    }
  }

  network_spec{
    network {
      name = "default"
    }
  }
}
