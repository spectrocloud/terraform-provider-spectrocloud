# Palette Dev Engine virtual cluster app tutorial example

End-to-end example accompanying the Spectro Cloud tutorial
[Deploy an Application using Palette Dev Engine](https://docs.spectrocloud.com/tutorials/pde/deploy-app/).

Palette Dev Engine (App Mode) lets you deploy applications onto lightweight virtual clusters carved out
of a shared cluster group, without provisioning or managing full Kubernetes clusters yourself. This
example deploys the Hello Universe sample application in two forms:
- **Scenario 1 (always deployed):** a virtual cluster running Hello Universe as a single, standalone
  UI container.
- **Scenario 2 (optional, via `enable-second-scenario`):** a second virtual cluster running the
  three-tier version of Hello Universe — UI, API, and a Postgres database, wired together through
  application-profile pack outputs.

It provisions, per scenario: a virtual cluster (`spectrocloud_virtual_cluster`), an application profile
describing the container/database packs (`spectrocloud_application_profile`), and the deployed application
itself (`spectrocloud_application`).

## Prerequisite: an existing cluster group

This example does not create a cluster group — it deploys virtual clusters into one that must already
exist. Create a cluster group first (via the Palette UI, under App Mode) before running this example, and
set `cluster-group-name` to its name.

## Other prerequisites

You will need the following before getting started:
1. A Palette API key, exported as the `SPECTROCLOUD_APIKEY` environment variable (or supplied via the
   `sc_api_key` variable).
2. An existing cluster group, as described above.

## Instructions

Clone this repository to a local directory, and change directory to
`examples/tutorials/pde-virtual-cluster-app`. Proceed with the following:
1. From the current directory, copy the template variable file `terraform.template.tfvars` to a new
   file `terraform.tfvars`.
2. Set `cluster-group-name` to your existing cluster group's name.
3. Initialize and run terraform: `terraform init && terraform apply`. This deploys scenario 1 (the
   standalone UI app).
4. To also deploy scenario 2 (the three-tier app), set `enable-second-scenario = true` and re-apply.

## Clean up

Run the destroy operation:

```shell
terraform destroy
```
