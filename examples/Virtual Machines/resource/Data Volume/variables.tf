variable "vm_name" {
  description = "Name of the existing spectrocloud_virtual_machine this data volume attaches to."
  default     = "tf-test-vm-basic-type"
}

variable "vm_namespace" {
  description = "Kubernetes namespace of the virtual machine."
  default     = "default"
}
