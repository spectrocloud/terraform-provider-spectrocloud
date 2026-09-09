# Tutorial-aligned examples

This directory holds Terraform examples that mirror specific tutorials published at
[docs.spectrocloud.com/tutorials](https://docs.spectrocloud.com/tutorials/). Unlike the general-purpose
examples under [`../resources`](../resources), [`../data-sources`](../data-sources), and [`../e2e`](../e2e),
each folder here is scoped to a single named tutorial, so you can follow the docs page and this code side
by side.

Each example folder contains its own `README.md` linking back to the docs tutorial it accompanies, along
with prerequisites and apply/destroy instructions specific to that example. New examples are ported
following [PORTING_CHECKLIST.md](PORTING_CHECKLIST.md).

| Folder | Docs tutorial | Status |
|---|---|---|
| [`custom-pack/`](custom-pack) | [Deploy a Custom Pack](https://docs.spectrocloud.com/tutorials/packs-registries/deploy-pack/) | Done |
| [`getting-started-multicloud/`](getting-started-multicloud) | [Cluster Management with Terraform (Getting Started)](https://docs.spectrocloud.com/tutorials/getting-started/palette/aws/deploy-manage-k8s-cluster-tf/) | Done |
| [`profile-variables/`](profile-variables) | [Deploy Cluster with Profile Variables](https://docs.spectrocloud.com/tutorials/profiles/cluster-profile-variables/) | Planned |
| [`cluster-templates/`](cluster-templates) | [Standardize Clusters with Cluster Templates (Terraform)](https://docs.spectrocloud.com/tutorials/clusters/cluster-templates/standardize-clusters-with-cluster-templates-terraform/) | Planned |
| [`pcg-vmware-app/`](pcg-vmware-app) | [Deploy App Workloads with a PCG](https://docs.spectrocloud.com/tutorials/clusters/pcg/deploy-app-pcg/) | Planned |
| [`pde-virtual-cluster-app/`](pde-virtual-cluster-app) | [Deploy an Application using Palette Dev Engine](https://docs.spectrocloud.com/tutorials/pde/deploy-app/) | Planned |
| [`cluster-profile-update/`](cluster-profile-update) | [Deploy Cluster Profile Updates](https://docs.spectrocloud.com/tutorials/profiles/update-k8s-cluster/) | Planned |

This table will be kept current as each example is ported and validated against the current provider build.
