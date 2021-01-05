# Basic Cluster demo

End-to-end example of provisioning a new GCP K8s cluster with all of its dependencies. This terraform configuration
will provision the following resources on Spectro Cloud:
- GCP Cloud Account
- GCP Cluster Profile
- GCP Cluster

## Instructions:

Clone this repository to a local directory, and change directory to `examples/e2e/gcp`. Proceed with the following:
1. Follow the Spectro Cloud documentations to create a GCP service account with appropriate permissions:
[GCP Service Account](https://docs.spectrocloud.com/clusters/?clusterType=google_cloud_cluster#creatingagcpcloudaccount).
2. From the current directory, copy the template variable file `terraform.template.tfvars` to a new file `terraform.tfvars`.
3. Specify and update all the placeholder values in the `terraform.tfvars` file.
4. Initialize and run terraform: `terraform init && terraform apply`.
5. Wait for the cluster creation to finish.

Once the cluster is provisioned, the cluster _kubeconfig_ file is exported in the current working directly.

Export the kubeconfig and check cluster pods:

```shell
export KUBECONFIG=kubeconfig_gcp-2
kubectl get pod -A
```

## Cleanup:

Run the destroy operation:

```shell
terraform destroy
```
