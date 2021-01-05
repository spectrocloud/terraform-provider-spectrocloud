# Basic Cluster demo

End-to-end example of provisioning a new VMware K8s cluster with all of its dependencies. This terraform configuration
will provision the following resources on Spectro Cloud:
- VMware Cluster Profile
- VMware Cluster

For SaaS deployments, there is a dependency to install the Private Cloud Gateway before VMware K8s clusters can be provisioned.
Please proceed with installation of this appliance by following the documentation:
[Spectro Cloud VMware Cluster](https://docs.spectrocloud.com/getting-started/?getting_started=vmware#yourfirstvmwarecluster).

Once the private cloud gateway is running, please note the name of the created vSphere cloud account. Cloud accounts are managed
in Spectro Cloud Admin view:
- Login to Spectro Cloud UI (e.g for SaaS: https://console.spectrocloud.com)
- Switch to the Admin view (bottom left in the main-menu)
- Select _Settings_ in the main-menu
- Select _Cloud Accounts_
- Note the name of the newly added vSphere cloud account

Alternatively, look at using a `spectrocloud_cloud_account_vsphere` resource to have Terraform create
a dedicated cloud account for this e2e example.

## Instructions:

Clone this repository to a local directory, and change directory to `examples/e2e/vsphere`. Proceed with the following:
1. Follow Spectro Cloud documentation to create the private cloud gateway:
[VMware First Cluster](https://docs.spectrocloud.com/getting-started/?getting_started=vmware#yourfirstvmwarecluster).
2. From the current directory, copy the template variable file `terraform.template.tfvars` to a new file `terraform.tfvars`.
3. Specify and update all the placeholder values in the `terraform.tfvars` file.
4. Initialize and run terraform: `terraform init && terraform apply`.
5. Wait for the cluster creation to finish.

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
