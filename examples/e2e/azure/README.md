# Basic Cluster demo

End-to-end example of provisioning a new Azure K8s cluster. This will provision a new cluster profile, cloud account, and cluster.

## Instructions:

Copy the template `terraform.template.tfvars` file to `terraform.tfvars`; and populate all placeholders.

Run with a command like this:

```shell
terraform init
terraform apply
```

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
