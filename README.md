![release](https://github.com/spectrocloud/terraform-provider-spectrocloud/workflows/release/badge.svg)
![ci](https://github.com/spectrocloud/terraform-provider-spectrocloud/actions/workflows/ci.yml/badge.svg)
![license](https://img.shields.io/github/license/spectrocloud/terraform-provider-spectrocloud)

# Terraform Provider for Spectro Cloud

Manage [Palette](https://www.spectrocloud.com/) and Palette VerteX — SaaS or on-prem — as infrastructure as code:
cloud accounts, cluster profiles, clusters, and more.

## Pre-requisites

- A Spectro Cloud account ([sign up for a free trial](https://www.spectrocloud.com/free-trial/))
- Terraform 0.13+
- kubectl 1.16+ (for interacting with provisioned clusters)

## Quick start

```hcl
terraform {
  required_providers {
    spectrocloud = {
      source  = "spectrocloud/spectrocloud"
      version = ">= 0.1"
    }
  }
}

provider "spectrocloud" {
  host    = var.sc_host
  api_key = var.sc_api_key
}
```

```console
terraform init && terraform apply
```

## Documentation & examples

| Looking for... | Go to |
|---|---|
| A searchable site with every example + full resource/data source reference | [`docsite/`](docsite/README.md) — clone this repo and open `docsite/output/index.html` (no install, no server) |
| Examples browsable right here on GitHub, by category | [`examples/`](examples/README.md) |
| End-to-end use cases (AWS, Azure, GCP, vSphere, MAAS, brownfield import, ...) | [`examples/end-to-end-usecases/`](examples/end-to-end-usecases/) |
| Tutorials aligned with docs.spectrocloud.com | [`examples/tutorials/`](examples/tutorials/) |
| Official schema reference (same content as the Registry) | [Terraform Registry docs](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs) |

## Develop

- Hack away
- Run `go generate` after your final commit
- If you touched `examples/` or `docs/resources` / `docs/data-sources`, regenerate the docs site too:
  `python3 docsite/generate_site.py` (see [`docsite/README.md`](docsite/README.md))
- Send in a PR

### Documentation conventions

Resource/data source documentation lives in [`docs/`](docs) and is generated via `tfplugindocs` from the templates in
[`templates/`](templates) — follow the Terraform Registry [documentation guidance](https://developer.hashicorp.com/terraform/registry/providers/docs)
and preview changes with the [Terraform Registry Preview Tool](https://registry.terraform.io/tools/doc-preview).

## Support

For questions or issues with the provider, open a discussion on the [discussion board](https://github.com/spectrocloud/terraform-provider-spectrocloud/discussions).
