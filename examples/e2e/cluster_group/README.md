# Virtual Cluster demo

End-to-end example of provisioning a Palette Virtual Cluster with all of its dependencies. This terraform configuration will provision the following resources on Spectro Cloud:
- K8s host cluster and datasource.
- Addon Cluster Profile

## Instructions:

Clone this repository to a local directory, and then change the directory to `examples/e2e/virtual`. Proceed with the following steps:
1. Provision host cluster as a prerequisite to use virtual cluster. 
To achieve it add host configuration block to existing cluster or provision it from scratch using one of existing examples:
<pre>
  host_config {
    host_endpoint_type = "LoadBalancer" 
    ingress_host       = "*.dev.spectrocloud.com"
  }
</pre>
2. From the current directory, copy the template variable file `terraform.template.tfvars` to a new file with the name `terraform.tfvars`.
3. Specify and update all the placeholder values in the `terraform.tfvars` file.
4. Initialize Terraform and invoke the deployment with the following command: `terraform init && terraform apply --auto-approve`.
5. Wait for the cluster creation to finish.

Once the cluster is provisioned, the cluster _kubeconfig_ file is exported in the current working directly.

Export the kubeconfig and check cluster pods:

```shell
export KUBECONFIG=kubeconfig_eks
kubectl get pod -A
```

## Cleanup:

Run the destroy operation:

```shell
terraform destroy
```
