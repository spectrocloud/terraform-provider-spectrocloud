# Basic Cluster demo

End-to-end example of provisioning a new VMware K8s cluster with all of its dependencies.

For SaaS deployments, there is a dependency to install the Private Cloud Gateway before VMware K8s clusters can be provisioned.
Please proceed with installation of this applicance by following the documentation:
[Spectro Cloud VMware Cluster](https://docs.spectrocloud.com/getting-started/?getting_started=vmware#yourfirstvmwarecluster).

Once the private cloud gateway is running, please note the name of the shared vsphere cloud account as `shared_vmware_cloud_account_name`.

Alternatively, look at using a `spectrocloud_cloud_account_vsphere` resource to create a dedicated cloud account for this cluster.

## Instructions:

Clone this repository to a local directory, and change directory to `examples/e2e/vsphere`. Proceed with the following:
1. Follow Spectro Cloud documentation to create the private cloud gateway:
[VMware First Cluster](https://docs.spectrocloud.com/getting-started/?getting_started=vmware#yourfirstvmwarecluster).
1. Copy the template variable file `terraform.template.tfvars` to `terraform.tfvars`.
1. Specify and upate all the placeholder values in the `terraform.tfvars` file.
1. Initialize and run terraform: `terraform init && terraform apply`.
1. Wait for the cluster creation to finish.

Once the cluster is provisioned, the cluster _kubeconfig_ file is exported in the current working directly.

Export the kubeconfig and check cluster pods:

```shell
export KUBECONFIG=kubeconfig_vsphere-2
kubectl get pod -A
```

## Cleanup:

Run the destroy operation:

```shell
terraform destroy
```
