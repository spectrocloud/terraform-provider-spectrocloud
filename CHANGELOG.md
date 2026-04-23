## 0.1.0 (Unreleased)

BACKWARDS INCOMPATIBILITIES / NOTES:

FEATURES:

* `resource/spectrocloud_cluster_maas`: Add `ssh_key` (deprecated) and `ssh_keys` attributes to the `cloud_config` block for SSH public key injection into MAAS nodes (`spectro` user). `ssh_key` and `ssh_keys` are mutually exclusive via `ConflictsWith`. Requires Palette with MAAS SSH key injection support for keys to be applied to running nodes (PCP-5897).
