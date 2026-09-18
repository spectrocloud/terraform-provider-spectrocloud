# PCG VMware app deployment tutorial example

End-to-end example accompanying the Spectro Cloud tutorial
[Deploy App Workloads with a PCG](https://docs.spectrocloud.com/tutorials/clusters/pcg/deploy-app-pcg/).

This terraform configuration provisions a VMware vSphere Kubernetes cluster, reachable through an
existing Private Cloud Gateway (PCG), with a cluster profile that includes MetalLB (for load-balanced
services on-prem) and the "Hello Universe" sample application. It will provision the following resources
on Spectro Cloud:
- Cluster Profile (Ubuntu, Kubernetes, Calico CNI, vSphere CSI, MetalLB, plus Hello Universe)
- VMware vSphere Kubernetes Cluster, deployed through the PCG

## Prerequisite: install the Private Cloud Gateway first

**Terraform cannot install the PCG itself** — this is a manual, one-time step you must complete before
running this example. Install a VMware PCG by following
[Spectro Cloud VMware Cluster](https://docs.spectrocloud.com/getting-started/?getting_started=vmware#yourfirstvmwarecluster).

Once the PCG is running, note the name of the vSphere cloud account it registered — this is a required
Terraform variable (`pcg_name`). Find it in the Palette UI: switch to the Admin view (bottom left of the
main menu) → Settings → Cloud Accounts.

## Other prerequisites

You will need the following before getting started:
1. A Palette API key, exported as the `SPECTROCLOUD_APIKEY` environment variable (or supplied via the
   `sc_api_key` variable).
2. The VMware PCG installed as described above, and its vSphere cloud account name.
3. A public SSH key to access the cluster nodes. If not provided, a new key pair is generated for you.

## Instructions

Clone this repository to a local directory, and change directory to
`examples/tutorials/pcg-vmware-app`. Proceed with the following:
1. From the current directory, copy the template variable file `terraform.template.tfvars` to a new
   file `terraform.tfvars`.
2. Specify and update all the placeholder (`REPLACE ME`) values in the `terraform.tfvars` file.
3. Initialize and run terraform: `terraform init && terraform apply`.
4. Wait for the cluster creation to finish.

If you didn't supply your own SSH key, see the `ssh_key_location` / `ssh_connection_command` outputs
for how to access the cluster nodes with the key Terraform generated for you.

## Clean up

Run the destroy operation:

```shell
terraform destroy
```
