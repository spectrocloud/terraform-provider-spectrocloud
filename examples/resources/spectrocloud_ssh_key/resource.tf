


resource "spectrocloud_ssh_key" "testKey_project"{
  name = "test-ssh-key-tf-project"
  ssh_key = "ssh-rsa AAAA6IEQhI1QLiicHLO5a== teerf2021"
  context = "project"
}

resource "spectrocloud_ssh_key" "testKey_tenant"{
  name = "test-ssh-key-tf-tenant"
  ssh_key = "ssh-rsa AAAA6IEQhI1QLiicHLO5a== teerf2021"
  context = "tenant"
}

data "spectrocloud_ssh_key" "sourceKey"{
  depends_on = [spectrocloud_ssh_key.testKey_project]
  name = "test-ssh-key-tf-project"
  context = "project"
}

output "ssh_key_id" {
  value = data.spectrocloud_ssh_key.sourceKey.id
}