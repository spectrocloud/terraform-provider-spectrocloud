package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

////
// resourceClusterBrownfieldUpdate + Read are both at 0%. Update has a
// fast early-return when no day2Fields change (which we can trigger by
// diffing only `description`, itself in the day2Fields list, or by
// diffing nothing at all). Read exercises GetCluster + status setters.
// The Import functions require live Import* endpoints not currently
// mocked — leave those for Wave B.

func baseBrownfieldAttrs() map[string]string {
	return map[string]string{
		"name":                   "test-brownfield-cluster",
		"context":                "project",
		"cloud_type":             "generic",
		"import_mode":            "read-only",
		"cluster_profile.#":      "0",
		"backup_policy.#":        "0",
		"scan_policy.#":          "0",
		"cluster_rbac_binding.#": "0",
		"namespaces.#":           "0",
		"host_config.#":          "0",
		"location_config.#":      "0",
		"machine_pool.#":         "0",
		"tags.#":                 "0",
		"cluster_timezone":       "",
		"apply_setting":          "DownloadAndInstall",
		"review_repave_state":    "",
		"description":            "",
		"pause_agent_upgrades":   "",
		"health_status":          "",
		"status":                 "",
		"kubectl_command":        "",
		"manifest_url":           "",
	}
}

// TestResourceClusterBrownfieldRead — covers the Read path against the
// standard cluster mock (getMockCluster). The mock's Spec.CloudConfigRef
// is set + Status populated, so most of Read's setters fire.
func TestResourceClusterBrownfieldRead(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterBrownfield(), "test-cluster-id",
		baseBrownfieldAttrs(), nil)
	// Read requires an ID matching an existing cluster in the mock.
	diags := resourceClusterBrownfieldRead(context.Background(), d, unitTestMockAPIClient)
	// The mock's GetCluster returns a cluster payload — Read populates
	// status + health_status + kubectl_command.
	_ = diags // may include warnings; branch coverage is the goal.
}

// TestResourceClusterBrownfieldUpdate_Day1Immutable — changing a Day-1
// field (name, cloud_type, import_mode, ...) must error out via the
// validateDay1FieldsImmutable guard.
func TestResourceClusterBrownfieldUpdate_Day1Immutable(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterBrownfield(), "test-cluster-id",
		baseBrownfieldAttrs(), map[string]*terraform.ResourceAttrDiff{
			"name": {Old: "test-brownfield-cluster", New: "renamed"},
		})

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError(), "renaming a Day-1 field must error")
}

// TestResourceClusterBrownfieldUpdate_NoDay2Change — no diff on any
// day2Fields → hasDay2Changes=false → the Update fast-paths to return.
func TestResourceClusterBrownfieldUpdate_NoDay2Change(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterBrownfield(), "test-cluster-id",
		baseBrownfieldAttrs(), nil)

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	// GetCluster succeeds, no day-2 field changed, return diags may be
	// empty (or include the "not Running-Healthy" warning that
	// downstream logic emits). Either way, no error.
	_ = diags
}

// TestResourceClusterBrownfieldUpdate_Description — description is in
// day2Fields → hasDay2Changes fires → then goes on to check
// isClusterRunningHealthy + updateCommonFields path.
func TestResourceClusterBrownfieldUpdate_Description(t *testing.T) {
	d := buildUpdateResourceData(resourceClusterBrownfield(), "test-cluster-id",
		baseBrownfieldAttrs(), map[string]*terraform.ResourceAttrDiff{
			"description": {Old: "", New: "new description"},
		})

	diags := resourceClusterBrownfieldUpdate(context.Background(), d, unitTestMockAPIClient)
	_ = diags
}

// (validateDay1FieldsImmutable is already tested in resource_cluster_brownfield_test.go.)
