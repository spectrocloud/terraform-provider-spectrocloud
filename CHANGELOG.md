## 0.1.0 (Unreleased)

BACKWARDS INCOMPATIBILITIES / NOTES:

* `resource/spectrocloud_cluster_maas`: Remove deprecated `cloud_config.ssh_key`. Configure SSH public keys with `cloud_config.ssh_keys` only.

FEATURES:

* `resource/spectrocloud_cluster_maas`: Add `ssh_keys` attribute to the `cloud_config` block for SSH public key injection into MAAS nodes (`spectro` user). Requires Palette with MAAS SSH key injection support for keys to be applied to running nodes (PCP-5897).
* `resource/spectrocloud_rate_config`: Add support for managing the tenant cloud rate config, the unit prices Palette uses to estimate cluster cloud cost and usage cost (PLT-2439).
* `data-source/spectrocloud_rate_config`: Add a data source for reading the tenant cloud rate config (PLT-2439).
