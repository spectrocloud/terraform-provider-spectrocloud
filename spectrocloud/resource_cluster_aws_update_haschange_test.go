package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// Update-HasChange suite test.
//
// resourceClusterAwsUpdate calls updateCommonFields, which has 12
// HasChange-guarded branches. Because schema.TestResourceData doesn't
// produce a diff, all 12 short-circuit under the current wave-2 tests,
// leaving the Update function at 23.9%.
//
// Using the InstanceState + InstanceDiff pattern (see
// cluster_update_haschange_test_helpers_test.go) lets HasChange fire
// for real. We diff scalar keys (not TypeList/TypeSet blocks) so the
// helpers don't try to unpack a partial block — the goal is to cover
// the outer HasChange branch + entry into the update helper. When the
// helper does an SDK call, it either succeeds (mocked) or errors —
// updateCommonFields bails out at the first error, but the branch is
// already covered.

// baseAwsUpdateAttrs is the "old" state — the minimum populated schema
// needed for validateSystemRepaveApproval → GetCluster → GetCloudConfigAws
// to succeed. Values match the mock cluster + AWS cloud config fixtures.
func baseAwsUpdateAttrs() map[string]string {
	return map[string]string{
		"name":                       "test-aws-cluster",
		"context":                    "project",
		"cloud_account_id":           "test-aws-account-id-1",
		"cloud_config_id":            "test-cloud-config-id",
		"cluster_profile.#":          "0",
		"cloud_config.#":             "0",
		"machine_pool.#":             "0",
		"cluster_rbac_binding.#":     "0",
		"namespaces.#":               "0",
		"host_config.#":              "0",
		"location_config.#":          "0",
		"scan_policy.#":              "0",
		"backup_policy.#":            "0",
		"cluster_meta_attribute":     "",
		"cluster_timezone":           "",
		"pause_agent_upgrades":       "",
		"os_patch_on_boot":           "false",
		"os_patch_schedule":          "",
		"os_patch_after":             "",
		"cluster_type":               "",
		"review_repave_state":        "",
		"renew_k8s_certificates_now": "false",
	}
}

// TestResourceClusterAwsUpdate_HasChange_ClusterTypeRejected — the outer
// validator ValidateClusterTypeUpdate rejects any change to cluster_type
// with a clear error. This test pins that guard.
func TestResourceClusterAwsUpdate_HasChange_ClusterTypeRejected(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		baseAwsUpdateAttrs(),
		map[string]*terraform.ResourceAttrDiff{
			"cluster_type": {Old: "", New: "changed"},
		})

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError(), "changing cluster_type must produce a validation error")
}

// TestResourceClusterAwsUpdate_HasChange_UpdateWorkerPoolsInParallelRejected —
// updateCommonFields rejects day-2 changes to update_worker_pools_in_parallel.
func TestResourceClusterAwsUpdate_HasChange_UpdateWorkerPoolsInParallelRejected(t *testing.T) {
	base := baseAwsUpdateAttrs()
	base["update_worker_pools_in_parallel"] = "false"

	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		base,
		map[string]*terraform.ResourceAttrDiff{
			"update_worker_pools_in_parallel": {Old: "false", New: "true"},
		})

	diags := resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError(), "changing update_worker_pools_in_parallel must produce a validation error")
	assert.Contains(t, diags[0].Summary, "update_worker_pools_in_parallel cannot be modified after cluster creation")
}

// TestResourceClusterAwsUpdate_HasChange_Timezone fires the
// cluster_timezone branch → updateClusterTimezone. When the "new"
// timezone is empty (or context is invalid) the helper's SDK call is
// skipped and it returns nil.
func TestResourceClusterAwsUpdate_HasChange_Timezone(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		baseAwsUpdateAttrs(),
		map[string]*terraform.ResourceAttrDiff{
			"cluster_timezone": {Old: "", New: ""},
		})
	_ = resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
}

// TestResourceClusterAwsUpdate_HasChange_ClusterMetaAttribute fires the
// cluster_meta_attribute branch → updateClusterAdditionalMetadata → SDK.
func TestResourceClusterAwsUpdate_HasChange_ClusterMetaAttribute(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		baseAwsUpdateAttrs(),
		map[string]*terraform.ResourceAttrDiff{
			"cluster_meta_attribute": {Old: "", New: "new-meta"},
		})
	_ = resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
}

// TestResourceClusterAwsUpdate_HasChange_PauseAgentUpgrades fires the
// pause_agent_upgrades branch → updateAgentUpgradeSetting → PUT
// /v1/spectroclusters/{uid}/upgrade/settings (mocked).
func TestResourceClusterAwsUpdate_HasChange_PauseAgentUpgrades(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		baseAwsUpdateAttrs(),
		map[string]*terraform.ResourceAttrDiff{
			"pause_agent_upgrades": {Old: "", New: "lock"},
		})
	_ = resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
}

// TestResourceClusterAwsUpdate_HasChange_OsPatchOnBoot fires the
// os_patch_on_boot branch → updateClusterOsPatchConfig → SDK.
func TestResourceClusterAwsUpdate_HasChange_OsPatchOnBoot(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		baseAwsUpdateAttrs(),
		map[string]*terraform.ResourceAttrDiff{
			"os_patch_on_boot": {Old: "false", New: "true"},
		})
	_ = resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
}

// TestResourceClusterAwsUpdate_HasChange_NameMetadata fires the
// name-tags-description branch → updateClusterMetadata → SDK.
func TestResourceClusterAwsUpdate_HasChange_NameMetadata(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		baseAwsUpdateAttrs(),
		map[string]*terraform.ResourceAttrDiff{
			"name": {Old: "test-aws-cluster", New: "renamed"},
		})
	_ = resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
}

// TestResourceClusterAwsUpdate_HasChange_RenewK8sCerts fires the final
// renewK8sCertificatesNow branch inside updateCommonFields.
func TestResourceClusterAwsUpdate_HasChange_RenewK8sCerts(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterAws(), "test-cluster-uid",
		baseAwsUpdateAttrs(),
		map[string]*terraform.ResourceAttrDiff{
			"renew_k8s_certificates_now": {Old: "false", New: "true"},
		})
	_ = resourceClusterAwsUpdate(context.Background(), d, unitTestMockAPIClient)
}
