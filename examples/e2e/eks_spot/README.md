# Basic Cluster demo

This example focused on creating EKS cluster with spot instances.
It is end-to-end example of provisioning a new EKS cluster with all of its dependencies. This terraform configuration
will provision the following resources on Spectro Cloud:
- AWS Cloud Account
- Eks Cluster Profile
- Eks Cluster

## Instructions:

Clone this repository to a local directory, and then change directory to `examples/e2e/eks_spot`. Proceed with the following:
1. Follow the Spectro Cloud documentation to create an AWS cloud account with appropriate permissions:
[AWS Cloud Account](https://docs.spectrocloud.com/clusters/?clusterType=aws_cluster#awscloudaccountpermissions).
2. From the current directory, copy the template variable file `terraform.template.tfvars` to a new file `terraform.tfvars`.
3. Specify and update all the placeholder values in the `terraform.tfvars` file.
4. Initialize and run terraform: `terraform init && terraform apply`.
5. Wait for the cluster creation to finish.

Once the cluster is provisioned, the cluster _kubeconfig_ file is exported in the current working directly.

Export the kubeconfig and check cluster pods:

```shell
export KUBECONFIG=kubeconfig_eks
kubectl get pod -A
```

## Clean up:

Run the destroy operation:

```shell
terraform destroy
```
