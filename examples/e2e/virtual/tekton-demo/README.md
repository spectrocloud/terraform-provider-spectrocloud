# Steps
1. Deploy Host Cluster on Palette
2. Enable Virtual Clusters & [configure Ingress for accessibility](https://docs.spectrocloud.com/clusters/sandbox-clusters/cluster-quickstart#ingress)
3. Fork https://github.com/TylerGillson/ulmaceae and clone it locally
4. Fill in `terraform.tfvars.template` & rename as `terraform.tfvars`
  - Note: use the Host DNS Pattern from Step 2 as `external_domain`
5. Run `generate_cosign_secret.sh` from within the `tekton-demo` directory
6. Run `terraform apply -auto-approve`
7. Once the virtual cluster deployment completes, trigger a Tekton pipeline to build an image from your demo repo and deploy it within the virtual cluster:
   ```bash
   git commit -a -m "trigger tekton pipeline" --allow-empty && git push origin master
   ```
8. Run `validate_taskrun.sh` to validate the latest TaskRun using cosign