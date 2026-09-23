# Getting started multi-cloud deployment tutorial example

End-to-end example accompanying the Spectro Cloud Getting Started tutorials:
[Cluster Management with Terraform](https://docs.spectrocloud.com/tutorials/getting-started/palette/aws/deploy-manage-k8s-cluster-tf/)
(the same Terraform pattern is used across the AWS, Azure, GCP, and VMware variants of this tutorial
under Getting Started → Palette).

This single Terraform configuration can deploy a Kubernetes cluster to any combination of AWS, Azure,
GCP, and VMware vSphere, each with a cluster profile that includes the "Hello Universe" sample application
and an optional Kubecost add-on. It will provision, per enabled cloud:
- A Cluster Profile (OS, Kubernetes, CNI, CSI layers, plus Hello Universe and optionally Kubecost)
- A Kubernetes Cluster on that cloud

## Choosing which cloud(s) to deploy to

The Terraform code has four main toggle variables, each independent — enable as many as you want, as
long as you provide the required values for each:

| Variable | Provider | Description | Default |
|---|---|---|---|
| `deploy-aws` | AWS | Enable to deploy a cluster to AWS. | `false` |
| `deploy-azure` | Azure | Enable to deploy a cluster to Azure. | `false` |
| `deploy-gcp` | GCP | Enable to deploy a cluster to GCP. | `false` |
| `deploy-vmware` | VMware vSphere | Enable to deploy a cluster to VMware vSphere. | `false` |

Each cloud also has a matching `deploy-<cloud>-kubecost` toggle that, when enabled, adds the Kubecost
add-on pack to that cloud's cluster profile (deployed as profile version `1.1.0` instead of `1.0.0`).

VMware additionally supports static IP placement via `deploy-vmware-static` (see the PCG-related variables
in `terraform.template.tfvars` if you enable it).

## Prerequisites

You will need the following before getting started:
1. A Palette API key, exported as the `SPECTROCLOUD_APIKEY` environment variable (or supplied via the
   `sc_api_key` variable).
2. A cloud account already registered in your Palette project settings for each cloud you enable.
3. For AWS: an existing EC2 key pair in the region you deploy to.
4. For VMware: a Private Cloud Gateway (PCG) already installed — see
   [Spectro Cloud VMware Cluster](https://docs.spectrocloud.com/getting-started/?getting_started=vmware#yourfirstvmwarecluster)
   — and its cloud account name.
5. Terraform 1.9 or later.

## Instructions

Clone this repository to a local directory, and change directory to
`examples/tutorials/getting-started-multicloud`. Proceed with the following:
1. From the current directory, copy the template variable file `terraform.template.tfvars` to a new
   file `terraform.tfvars`.
2. Set the `deploy-*` toggle(s) for the cloud(s) you want, and fill in all the placeholder
   (`REPLACE ME`) values relevant to those clouds. Values for clouds you leave disabled don't need to
   be filled in.
3. Initialize and run terraform: `terraform init && terraform apply`.
4. Wait for the cluster creation to finish.

Once a cluster is provisioned, its Hello Universe application becomes reachable through the cluster's
load balancer — see the `advisory` output for a note on DNS propagation timing. For VMware deployments
without a supplied SSH key, see the `ssh_key_location` / `ssh_connection_command` outputs for how to
access the nodes.

## Clean up

Run the destroy operation:

```shell
terraform destroy
```
