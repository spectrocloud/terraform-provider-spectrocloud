package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// Update-HasChange suite test.
//
// Uses the InstanceState+InstanceDiff pattern to fire HasChange
// branches inside resourceClusterAksUpdate / updateCommonFields. See
// resource_cluster_aws_update_haschange_test.go for the rationale.

func baseAksUpdateAttrs() map[string]string {
	return map[string]string{
		"name":                       "test-aks-cluster",
		"context":                    "project",
		"cloud_account_id":           "test-azure-account-id-1",
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

// runUpdateHasChangeSuite runs a compact set of "safe" scalar-key diffs
// against any cluster resource's Update func. Each diff fires one
// HasChange branch inside updateCommonFields.
//
// Callers pass the resource factory, the ID to seed on state, and the
// baseline attrs. This lets Wave A avoid ~120 lines of boilerplate per
// cluster type. rejectsClusterType tells the suite whether the target
// resource wires ValidateClusterTypeUpdate — only AWS / MaaS /
// custom_cloud do.
func runUpdateHasChangeSuite(
	t *testing.T,
	name string,
	res *schema.Resource,
	id string,
	base map[string]string,
	updater func(context.Context, *schema.ResourceData, interface{}) any,
	rejectsClusterType bool,
) {
	t.Helper()
	cases := []struct {
		label string
		diff  map[string]*terraform.ResourceAttrDiff
	}{
		{"Timezone", simpleDiff("cluster_timezone", "", "")},
		{"ClusterMetaAttribute", simpleDiff("cluster_meta_attribute", "", "new-meta")},
		{"PauseAgentUpgrades", simpleDiff("pause_agent_upgrades", "", "lock")},
		{"OsPatchOnBoot", simpleDiff("os_patch_on_boot", "false", "true")},
		{"NameMetadata", simpleDiff("name", base["name"], "renamed-cluster")},
		{"RenewK8sCerts", simpleDiff("renew_k8s_certificates_now", "false", "true")},
	}
	for _, c := range cases {
		c := c
		t.Run(name+"/"+c.label, func(t *testing.T) {
			d := buildUpdateResourceData(res, id, base, c.diff)
			_ = updater(context.Background(), d, unitTestMockAPIClient)
		})
	}

	if rejectsClusterType {
		t.Run(name+"/ClusterTypeRejected", func(t *testing.T) {
			d := buildUpdateResourceData(res, id, base, simpleDiff("cluster_type", "", "changed"))
			out := updater(context.Background(), d, unitTestMockAPIClient)
			if diags, ok := out.(diagsLike); ok {
				assert.True(t, diags.HasError(), "cluster_type change must error")
			}
		})
	}
}

// diagsLike lets us call HasError() without importing diag in the test
// helper — matches diag.Diagnostics's method set.
type diagsLike interface {
	HasError() bool
}

// TestResourceClusterAksUpdate_HasChange_Suite fires all common-fields
// HasChange branches through the AKS Update entry point.
func TestResourceClusterAksUpdate_HasChange_Suite(t *testing.T) {
	// resourceClusterAksUpdate returns diag.Diagnostics — wrap it so
	// the suite helper can invoke it through an interface{}-shaped
	// signature.
	updater := func(ctx context.Context, d *schema.ResourceData, m interface{}) any {
		return resourceClusterAksUpdate(ctx, d, m)
	}
	runUpdateHasChangeSuite(t, "AKS", resourceClusterAks(), "test-cluster-uid",
		baseAksUpdateAttrs(), updater, false)
}
