# Cluster profile update tutorial example

End-to-end example accompanying the Spectro Cloud tutorial
[Deploy Cluster Profile Updates](https://docs.spectrocloud.com/tutorials/profiles/update-k8s-cluster/).

This is a day-2 operations example: it deploys a two-cluster application (a Hello Universe frontend and
a separate hello-universe-api backend) per enabled cloud, then shows how to roll the frontend cluster
forward to a new cluster profile version — via Terraform, in place — once the backend's address is known.
It provisions, per enabled cloud (AWS, Azure, and/or GCP):
- A frontend cluster profile (version `1.0.0`: OS, Kubernetes, CNI, CSI, plus the standalone Hello
  Universe UI) and cluster
- A backend cluster profile (OS, Kubernetes, CNI, CSI, plus the hello-universe-api + database manifests)
  and cluster
- A local kubeconfig file for the backend cluster, so you can query its load balancer address

## The Day-2 workflow this demonstrates

1. **Initial deploy:** with `update_profile_version = false` (default), `terraform apply` deploys both
   the frontend cluster (profile v1.0.0, standalone UI) and the backend cluster for each enabled cloud.
2. **Find the backend's address:** export the generated kubeconfig and query the backend's load balancer,
   per the `<cloud>_hello_universe_api_ip` output shown after apply (for example
   `export KUBECONFIG=$(pwd)/aws-cluster-api.kubeconfig && kubectl get service hello-universe-api-service --namespace hello-universe-api -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'`).
3. **Point the frontend at the backend:** set the matching `<cloud>-hello-universe-api-uri` variable to
   that address.
4. **Roll the frontend forward:** set `update_profile_version = true` and re-apply. This creates cluster
   profile version `1.1.0` (the same packs, plus the Hello Universe manifest updated with an `API_URI`
   environment variable) and updates the frontend cluster in place to use it — no cluster recreation.

## Prerequisites

You will need the following before getting started:
1. A Palette API key, exported as the `SPECTROCLOUD_APIKEY` environment variable (or supplied via the
   `sc_api_key` variable).
2. A cloud account already registered in your Palette project settings for each cloud you enable.
3. For AWS: an existing EC2 key pair in the region you deploy to.
4. `kubectl` installed locally, to query the backend cluster's load balancer address.

## Instructions

Clone this repository to a local directory, and change directory to
`examples/tutorials/cluster-profile-update`. Proceed with the following:
1. From the current directory, copy the template variable file `terraform.template.tfvars` to a new
   file `terraform.tfvars`.
2. Set the `deploy-<cloud>` toggle(s) for the cloud(s) you want, and fill in the placeholder
   (`REPLACE_ME`) values relevant to those clouds.
3. Initialize and run terraform: `terraform init && terraform apply`.
4. Walk through the Day-2 workflow above to point the frontend at the backend and roll it forward.

## Clean up

Run the destroy operation:

```shell
terraform destroy
```
