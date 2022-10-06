########################################################################
# uncomment everything below this line to deploy the tekton demo stack #
########################################################################

# data "spectrocloud_registry" "registry" {
#   name = var.registry_name
# }

# data "spectrocloud_pack" "tekton-operator" {
#   name = "tekton-operator"
#   registry_uid = data.spectrocloud_registry.registry.id
#   version  = "0.61.0"
# }

# # REQUIRES K8s >= v1.22 !!!
# data "spectrocloud_pack" "tekton-chains" {
#   name = "tekton-chains"
#   registry_uid = data.spectrocloud_registry.registry.id
#   version  = "0.12.0"
# }

# locals {
#   pack_values_chains_ns = <<-EOT
#     pack:
#       namespace: tekton-chains
#   EOT
# }
# locals {
#   pack_values_demo_ns = <<-EOT
#     pack:
#       namespace: tekton-trigger-demo
#   EOT
# }

# resource "spectrocloud_cluster_profile" "profile" {
#   name        = "tekton-tf-stack"
#   description = "Tekton Pipelines, Triggers, and Chains"
#   cloud       = "all"
#   type        = "add-on"

#   pack {
#     name   = data.spectrocloud_pack.tekton-operator.name
#     tag    = data.spectrocloud_pack.tekton-operator.version
#     uid    = data.spectrocloud_pack.tekton-operator.id
#     values = data.spectrocloud_pack.tekton-operator.values
#   }

#   # put the cosign secret in its own pack so that it reconciles 
#   # cleanly on the first pass and exists before the chains operator is deployed
#   pack {
#     name = "tekton-cosign-setup"
#     type = "manifest"
#     values = local.pack_values_chains_ns

#     manifest {
#       name    = "secret-cosign"
#       content = file("${path.cwd}/tekton-demo/manifests/secret-cosign.yaml")
#     }
#   }

#   pack {
#     name   = data.spectrocloud_pack.tekton-chains.name
#     tag    = data.spectrocloud_pack.tekton-chains.version
#     uid    = data.spectrocloud_pack.tekton-chains.id
#     values = data.spectrocloud_pack.tekton-chains.values
#   }

#   pack {
#     name = "tekton-demo-setup"
#     type = "manifest"
#     values = local.pack_values_demo_ns

#     manifest {
#       name    = "rbac"
#       content = file("${path.cwd}/tekton-demo/manifests/rbac.yaml")
#     }
#     manifest {
#       name    = "secret-tekon"
#       content = replace(replace(
#         file("${path.cwd}/tekton-demo/manifests/secret-tekton.yaml"),
#         "<GITHUB_ACCESS_TOKEN>", var.github_access_token),
#         "<DOCKER_CONFIG>",var.docker_config,
#       )
#     }
#     manifest {
#       name    = "triggers"
#       content = replace(
#         file("${path.cwd}/tekton-demo/manifests/triggers.yaml"),
#         "<IMAGE_SOURCE>", var.image_source,
#       )
#     }
#     manifest {
#       name    = "pipeline-create-ingress"
#       content = replace(
#         file("${path.cwd}/tekton-demo/manifests/pipeline-create-ingress.yaml"),
#         "<EXTERNAL_DOMAIN>", var.external_domain,
#       )
#     }
#     manifest {
#       name    = "pipeline-create-webhook"
#       content = replace(replace(replace(replace(
#         file("${path.cwd}/tekton-demo/manifests/pipeline-create-webhook.yaml"),
#         "<EXTERNAL_DOMAIN>", var.external_domain),
#         "<GITHUB_ORG>", var.github_org),
#         "<GITHUB_REPO>", var.github_repo),
#         "<GITHUB_USER>", var.github_user,
#       )
#     }
#     manifest {
#       name    = "pipeline-build-deploy"
#       content = file("${path.cwd}/tekton-demo/manifests/pipeline-build-deploy.yaml")
#     }
#   }

#   pack {
#     name = "tekton-demo-pipelineruns"
#     type = "manifest"
#     values = local.pack_values_demo_ns

#     manifest {
#       name    = "pipelineruns"
#       content = file("${path.cwd}/tekton-demo/manifests/pipelineruns.yaml")
#     }
#   }
# }
