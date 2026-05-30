## 0.1.0 (Unreleased)

BACKWARDS INCOMPATIBILITIES / NOTES:

* `resource/spectrocloud_cluster_maas`: Remove deprecated `cloud_config.ssh_key`. Configure SSH public keys with `cloud_config.ssh_keys` only.

FEATURES:

* `resource/spectrocloud_cluster_maas`: Add `ssh_keys` attribute to the `cloud_config` block for SSH public key injection into MAAS nodes (`spectro` user). Requires Palette with MAAS SSH key injection support for keys to be applied to running nodes (PCP-5897).
