
data "spectrocloud_ssh_key" "sourceKey"{
  name = "spectro2021"
  context = "project"
}

resource "spectrocloud_ssh_key" "testKey_project"{
  name = "test-ssh-key-tf-project"
  ssh_key = base64decode(data.spectrocloud_ssh_key.sourceKey.ssh_key)
  context = "project"
}

resource "spectrocloud_ssh_key" "testKey_tenant"{
  name = "test-ssh-key-tf-tenant"
  ssh_key = "ssh-rsa AAAA6IEQhI1QLiicHLO5a== teerf2021"
  context = "tenant"
}