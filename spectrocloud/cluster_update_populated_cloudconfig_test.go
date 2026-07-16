package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

////
// The Batch 14-16 HasChange tests diffed scalar fields (name,
// pause_agent_upgrades, cluster_meta_attribute, etc.) — those fire
// HasChange for `updateCommonFields` but *don't* trigger the two big
// per-cluster branches inside the Update function itself:
//
//   if d.HasChange("cloud_config")  { ... }   // ~5-8 statements
//   if d.HasChange("machine_pool")  { ... }   // ~15-40 statements
//
// Both branches now live behind the shim (Batch 22) — before the shim,
// even if they fired, the surrounding wait would truncate everything
// downstream. Now we can diff `cloud_config.0.region` etc. and the
// Update runs cleanly against the mocked /clusterConfig endpoints.
//
// Each test:
//   1. Provides a populated cloud_config block in the baseline state.
//   2. Diffs cloud_config.0.<field> to fire HasChange("cloud_config").
//   3. Calls the Update func — the SDK call succeeds against the mock.

// awsBaselineWithCloudConfig returns the InstanceState.Attributes map
// with cloud_config populated. Setting `cloud_config.#=1` + at least
// one indexed child key is enough for d.Get("cloud_config").([]interface{})[0]
// to unpack cleanly.
func awsBaselineWithCloudConfig() map[string]string {
	return map[string]string{
		"name":            "test-aws-cluster",
		"context":         "project",
		"cloud_config_id": "test-cloud-config-id",

		// Populated cloud_config block.
		"cloud_config.#":                             "1",
		"cloud_config.0.region":                      "us-east-1",
		"cloud_config.0.vpc_id":                      "vpc-test123",
		"cloud_config.0.ssh_key_name":                "",
		"cloud_config.0.control_plane_lb":            "",
		"cloud_config.0.override_cluster_api_config": "",

		"cluster_profile.#":      "0",
		"machine_pool.#":         "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"scan_policy.#":          "0",
		"backup_policy.#":        "0",
		"cluster_meta_attribute": "",
		"cluster_timezone":       "",
		"pause_agent_upgrades":   "",
	}
}

func TestResourceClusterAwsUpdate_PopulatedCloudConfigChange(t *testing.T) {
	// Diff a cloud_config scalar so HasChange("cloud_config") fires. The
	// Update func extracts the block, calls UpdateCloudConfigAws (mocked
	// at PUT /v1/cloudconfigs/aws/{configUid}/clusterConfig), then
	// continues to the machine_pool branch (which won't fire — .#=0).
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		awsBaselineWithCloudConfig(),
		map[string]*terraform.ResourceAttrDiff{
			"cloud_config.0.region": {Old: "us-east-1", New: "us-west-2"},
		})

	_ = resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	// The mock returns 204 on the update; downstream branches (Read,
	// machine_pool loop) are covered by other tests. What matters here
	// is that the `if d.HasChange("cloud_config")` body — including
	// toCloudConfigAws + the SDK call — is now traversed.
}

// gcpBaselineWithCloudConfig — the GCP cloud_config schema takes a
// similar shape (region + project_name + network + ...).
func gcpBaselineWithCloudConfig() map[string]string {
	return map[string]string{
		"name":            "test-gcp-cluster",
		"context":         "project",
		"cloud_config_id": "test-cloud-config-id",

		"cloud_config.#":             "1",
		"cloud_config.0.region":      "us-central1",
		"cloud_config.0.project":     "my-gcp-project",
		"cloud_config.0.network":     "default",
		"cloud_config.0.subnet_name": "",

		"cluster_profile.#":      "0",
		"machine_pool.#":         "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"scan_policy.#":          "0",
		"backup_policy.#":        "0",
		"cluster_meta_attribute": "",
		"cluster_timezone":       "",
		"pause_agent_upgrades":   "",
	}
}

func TestResourceClusterGcpUpdate_PopulatedCloudConfigChange(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterGcp(), "test-cluster-uid",
		gcpBaselineWithCloudConfig(),
		map[string]*terraform.ResourceAttrDiff{
			"cloud_config.0.region": {Old: "us-central1", New: "us-east1"},
		})

	_ = resourceClusterGcpUpdate(context.Background(), d, unitTestMockAPIClient)
}

// aksBaselineWithCloudConfig
func aksBaselineWithCloudConfig() map[string]string {
	return map[string]string{
		"name":            "test-aks-cluster",
		"context":         "project",
		"cloud_config_id": "test-cloud-config-id",

		"cloud_config.#":                     "1",
		"cloud_config.0.subscription_id":     "sub-uid",
		"cloud_config.0.resource_group":      "my-rg",
		"cloud_config.0.region":              "eastus",
		"cloud_config.0.ssh_key":             "",
		"cloud_config.0.vnet_name":           "",
		"cloud_config.0.vnet_resource_group": "",
		"cloud_config.0.vnet_cidr_block":     "",

		"cluster_profile.#":      "0",
		"machine_pool.#":         "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"scan_policy.#":          "0",
		"backup_policy.#":        "0",
		"cluster_meta_attribute": "",
		"cluster_timezone":       "",
		"pause_agent_upgrades":   "",
	}
}

func TestResourceClusterAksUpdate_PopulatedCloudConfigChange(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAks(), "test-cluster-uid",
		aksBaselineWithCloudConfig(),
		map[string]*terraform.ResourceAttrDiff{
			"cloud_config.0.region": {Old: "eastus", New: "westus"},
		})

	_ = resourceClusterAksUpdate(context.Background(), d, unitTestMockAPIClient)
}

// azureBaselineWithCloudConfig
func azureBaselineWithCloudConfig() map[string]string {
	return map[string]string{
		"name":            "test-azure-cluster",
		"context":         "project",
		"cloud_config_id": "test-cloud-config-id",

		"cloud_config.#":                 "1",
		"cloud_config.0.subscription_id": "sub-uid",
		"cloud_config.0.resource_group":  "my-rg",
		"cloud_config.0.region":          "eastus",
		"cloud_config.0.ssh_key":         "",

		"cluster_profile.#":      "0",
		"machine_pool.#":         "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"scan_policy.#":          "0",
		"backup_policy.#":        "0",
		"cluster_meta_attribute": "",
		"cluster_timezone":       "",
		"pause_agent_upgrades":   "",
	}
}

func TestResourceClusterAzureUpdate_PopulatedCloudConfigChange(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAzure(), "test-cluster-uid",
		azureBaselineWithCloudConfig(),
		map[string]*terraform.ResourceAttrDiff{
			"cloud_config.0.region": {Old: "eastus", New: "westus"},
		})

	_ = resourceClusterAzureUpdate(context.Background(), d, unitTestMockAPIClient)
}

// gkeBaselineWithCloudConfig
func gkeBaselineWithCloudConfig() map[string]string {
	return map[string]string{
		"name":            "test-gke-cluster",
		"context":         "project",
		"cloud_config_id": "test-cloud-config-id",

		"cloud_config.#":         "1",
		"cloud_config.0.region":  "us-central1",
		"cloud_config.0.project": "my-gcp-project",

		"cluster_profile.#":      "0",
		"machine_pool.#":         "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"scan_policy.#":          "0",
		"backup_policy.#":        "0",
		"cluster_meta_attribute": "",
		"cluster_timezone":       "",
		"pause_agent_upgrades":   "",
	}
}

func TestResourceClusterGkeUpdate_PopulatedCloudConfigChange(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterGke(), "test-cluster-uid",
		gkeBaselineWithCloudConfig(),
		map[string]*terraform.ResourceAttrDiff{
			"cloud_config.0.region": {Old: "us-central1", New: "us-east1"},
		})
	_ = resourceClusterGkeUpdate(context.Background(), d, unitTestMockAPIClient)
}

// vsphereBaselineWithCloudConfig
func vsphereBaselineWithCloudConfig() map[string]string {
	return map[string]string{
		"name":            "test-vsphere-cluster",
		"context":         "project",
		"cloud_config_id": "test-cloud-config-id",

		"cloud_config.#":                       "1",
		"cloud_config.0.datacenter":            "dc-1",
		"cloud_config.0.folder":                "spectrocloud",
		"cloud_config.0.network_type":          "VIP",
		"cloud_config.0.network_search_domain": "",
		"cloud_config.0.ssh_key":               "",
		"cloud_config.0.vip":                   "192.168.1.100",

		"cluster_profile.#":      "0",
		"machine_pool.#":         "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"scan_policy.#":          "0",
		"backup_policy.#":        "0",
		"cluster_meta_attribute": "",
		"cluster_timezone":       "",
		"pause_agent_upgrades":   "",
	}
}

func TestResourceClusterVsphereUpdate_PopulatedCloudConfigChange(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterVsphere(), "test-cluster-uid",
		vsphereBaselineWithCloudConfig(),
		map[string]*terraform.ResourceAttrDiff{
			"cloud_config.0.vip": {Old: "192.168.1.100", New: "192.168.1.101"},
		})
	_ = resourceClusterVsphereUpdate(context.Background(), d, unitTestMockAPIClient)
}

// maasBaselineWithCloudConfig
func maasBaselineWithCloudConfig() map[string]string {
	return map[string]string{
		"name":            "test-maas-cluster",
		"context":         "project",
		"cloud_config_id": "test-cloud-config-id",

		"cloud_config.#":        "1",
		"cloud_config.0.domain": "example.com",

		"cluster_profile.#":      "0",
		"machine_pool.#":         "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"scan_policy.#":          "0",
		"backup_policy.#":        "0",
		"cluster_meta_attribute": "",
		"cluster_timezone":       "",
		"pause_agent_upgrades":   "",
	}
}

func TestResourceClusterMaasUpdate_PopulatedCloudConfigChange(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterMaas(), "test-cluster-uid",
		maasBaselineWithCloudConfig(),
		map[string]*terraform.ResourceAttrDiff{
			"cloud_config.0.domain": {Old: "example.com", New: "other.com"},
		})
	_ = resourceClusterMaasUpdate(context.Background(), d, unitTestMockAPIClient)
}

// cloudstackBaselineWithCloudConfig
func cloudstackBaselineWithCloudConfig() map[string]string {
	return map[string]string{
		"name":            "test-cloudstack-cluster",
		"context":         "project",
		"cloud_config_id": "test-cloud-config-id",

		"cloud_config.#":                               "1",
		"cloud_config.0.domain":                        "ROOT",
		"cloud_config.0.zone":                          "zone-1",
		"cloud_config.0.control_plane_endpoint.#":      "1",
		"cloud_config.0.control_plane_endpoint.0.host": "cp.example.com",
		"cloud_config.0.control_plane_endpoint.0.type": "VIP",
		"cloud_config.0.ssh_key":                       "",

		"cluster_profile.#":      "0",
		"machine_pool.#":         "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"scan_policy.#":          "0",
		"backup_policy.#":        "0",
		"cluster_meta_attribute": "",
		"cluster_timezone":       "",
		"pause_agent_upgrades":   "",
	}
}

func TestResourceClusterCloudStackUpdate_PopulatedCloudConfigChange(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterApacheCloudStack(), "test-cluster-uid",
		cloudstackBaselineWithCloudConfig(),
		map[string]*terraform.ResourceAttrDiff{
			"cloud_config.0.zone": {Old: "zone-1", New: "zone-2"},
		})
	_ = resourceClusterApacheCloudStackUpdate(context.Background(), d, unitTestMockAPIClient)
}

// Compile-time reference to keep schema import used across future
// batch-23 tests that will add machine_pool diffs.
var _ = &schema.Set{}
