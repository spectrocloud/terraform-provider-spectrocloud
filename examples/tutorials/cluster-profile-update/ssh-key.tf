resource "tls_private_key" "tutorial_ssh_key" {
  count     = var.deploy-azure ? 1 : 0
  algorithm = "RSA"
  rsa_bits  = "4096"
}
