resource "spectrocloud_cloudaccount_tencent" "account" {
  name               = "tencent-tke-tf1"
  tencent_secret_id  = var.tencent_secret_id
  tencent_secret_key = var.tencent_secret_key
}
