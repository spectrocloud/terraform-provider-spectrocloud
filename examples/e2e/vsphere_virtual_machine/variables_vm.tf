variable "name" {
  type = string
}

variable "namespace" {
  type = string
}

variable "state" {
  type    = string
  default = "Running"
}

variable "cpu_cores" {
  type    = number
  default = 1
}

variable "memory" {
  type    = string
  default = "2G"
}

variable "image" {
  type    = string
  default = "quay.io/kubevirt/alpine-container-disk-demo"
}

variable "cloudinit_user_data" {
  type    = string
  default = "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n            "
}
