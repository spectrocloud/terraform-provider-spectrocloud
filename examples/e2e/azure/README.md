# Basic Cluster demo

End-to-end example of provisioning a new Azure K8s cluster with all of its dependencies.

## Instructions:

Clone this repository to a local directory, and then change directory to `examples/e2e/azure`. Proceed with the following:
1. Follow the Spectro Cloud documentations for creating an Azure Application Registration:
[Azure Cloud Account](https://docs.spectrocloud.com/clusters?clusterType=azure_cluster#creatinganazurecloudaccount).
1. Copy the template variable file `terraform.template.tfvars` to `terraform.tfvars`.
1. Specify and upate all the placeholder values in the `terraform.tfvars` file.
1. Initialize and run terraform: `terraform init && terraform apply`.
1. Wait for the cluster creation to finish.

Once the cluster is provisioned, the cluster _kubeconfig_ file is exported in the current working directly.

Export the kubeconfig and check cluster pods:

```shell
export KUBECONFIG=kubeconfig_azure-2
kubectl get pod -A
```

## Cleanup:

Run the destroy operation:

```shell
terraform destroy
```
