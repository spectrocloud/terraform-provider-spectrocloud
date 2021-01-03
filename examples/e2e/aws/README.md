# Basic Cluster demo

End-to-end example of provisioning a new AWS K8s cluster with all of its dependencies.

## Instructions:

Clone this repository to a local directory, and then change directory to `examples/e2e/aws`. Proceed with the following:
1. Follow the Spectro Cloud documentations for to create an AWS cloud account with appropriate permissions:
[AWS Cloud Account](https://docs.spectrocloud.com/clusters/?clusterType=aws_cluster#awscloudaccountpermissions).
1. Copy the template variable file `terraform.template.tfvars` to `terraform.tfvars`.
1. Specify and upate all the placeholder values in the `terraform.tfvars` file.
1. Initialize and run terraform: `terraform init && terraform apply`.
1. Wait for the cluster creation to finish.

Once the cluster is provisioned, the cluster _kubeconfig_ file is exported in the current working directly.

Export the kubeconfig and check cluster pods:

```shell
export KUBECONFIG=kubeconfig_aws-2
kubectl get pod -A
```

## Cleanup:

Run the destroy operation:

```shell
terraform destroy
```
