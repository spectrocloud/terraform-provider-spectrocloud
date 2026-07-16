package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Update-HasChange suite test.
// Edge Native uses no dedicated cloud account UID (optional field), so
// the base attrs leave it empty. The Update func still needs the mock
// cluster + cloud config to be resolvable.

func baseEdgeNativeUpdateAttrs() map[string]string {
	return map[string]string{
		"name":                       "test-edge-native-cluster",
		"context":                    "project",
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
		"review_repave_state":        "",
		"renew_k8s_certificates_now": "false",
	}
}

func TestResourceClusterEdgeNativeUpdate_HasChange_Suite(t *testing.T) {
	updater := func(ctx context.Context, d *schema.ResourceData, m interface{}) any {
		return resourceClusterEdgeNativeUpdate(ctx, d, m)
	}
	runUpdateHasChangeSuite(t, "EdgeNative", resourceClusterEdgeNative(),
		"test-cluster-uid", baseEdgeNativeUpdateAttrs(), updater, false)
}
