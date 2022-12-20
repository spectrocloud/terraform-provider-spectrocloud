# Virtual Cluster Group Demo

This is an end-to-end example of provisioning a Palette Virtual Cluster Group with all of its dependencies. This Terraform configuration will provision the following resources on Spectro Cloud:
- K8s host cluster and datasource.
- Addon Cluster Profile

## Instructions

Clone this repository to a local directory, and then change the directory to `examples/e2e/virtual`. Proceed with the following steps:
1. Provision a host cluster. This is a prerequisite before using a Palette Virtual Cluster Group. 
To create a virtual cluster group, either add a host configuration block to an existing host cluster or provision a new host cluster from scratch using one of the existing examples:
<pre>
  host_config {
    host_endpoint_type = "LoadBalancer" 
    ingress_host       = "*.dev.spectrocloud.com"
  }
</pre>
2. From the current directory, copy the template variable file `terraform.template.tfvars` to a new file with the name `terraform.tfvars`.
3. Specify and update all the placeholder values in the `terraform.tfvars` file.
4. Initialize Terraform and invoke the deployment with the following command: `terraform init && terraform apply --auto-approve`.
5. Wait for the cluster group creation to finish.

## Clean up:

Run the destroy operation:

```shell
terraform destroy
```
