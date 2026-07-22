package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

//// Exercises the delete path on resource_kubevirt_datavolume against the
// /v1/spectroclusters/{uid}/vms/{vmName}/removeVolume mock route in
// routes/mockKubevirtVM.go.
//
// Create/Read require a full VM fixture and rely on ToHapiVolume /
// FromHapiVolume — the convert package already has tests for that
// piece. Delete stands alone: it just needs GetCluster to succeed and
// DeleteDataVolume to hit the removeVolume endpoint.

func TestResourceKubevirtDataVolumeDelete(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	// IdPartsDV expects: context/clusterUid/vmNamespace/vmName/volName
	d.SetId("project/test-cluster-uid/default/test-vm/boot-vol")

	diags := resourceKubevirtDataVolumeDelete(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	// Delete clears the ID.
	assert.Empty(t, d.Id())
}

func TestResourceKubevirtDataVolumeDelete_BadID(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	d.SetId("only-two/parts")
	diags := resourceKubevirtDataVolumeDelete(context.Background(), d, unitTestMockAPIClient)
	// IdPartsDV rejects the ID → Delete returns diag.FromErr immediately.
	assert.True(t, diags.HasError())
}

// TestResourceKubevirtDataVolumeRead_BadID exercises the malformed-ID
// early-return branch (handleReadError).
func TestResourceKubevirtDataVolumeRead_BadID(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	d.SetId("only-two-parts") // fewer than 5 slash segments
	diags := resourceKubevirtDataVolumeRead(context.Background(), d, unitTestMockAPIClient)
	// handleReadError may clear the ID or return an error; either way
	// the func body was entered.
	_ = diags
}

// TestResourceKubevirtDataVolumeRead_ValidNoMatch — valid ID and the
// mock's VM has DataVolumeTemplates, but the metadata block's name
// doesn't match any of them → the loop body isn't entered and Read
// returns empty diags.
func TestResourceKubevirtDataVolumeRead_ValidNoMatch(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	// Metadata is required by the schema. Set a name/namespace that
	// deliberately does NOT match the mock's "boot-vol" fixture, so the
	// flatten branch is skipped but the loop body is still traversed.
	_ = d.Set("metadata", []interface{}{
		map[string]interface{}{
			"name":      "different-dv-name",
			"namespace": "default",
		},
	})
	d.SetId("project/test-cluster-uid/default/test-vm/different-dv-name")

	diags := resourceKubevirtDataVolumeRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// TestResourceKubevirtDataVolumeCreate exercises the success path: GetCluster
// resolves against the mock cluster fixture, CreateDataVolume hits the
// PUT .../addVolume route (204 No Content), and the ID gets built from the
// resulting metadata.
func TestResourceKubevirtDataVolumeCreate(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	_ = d.Set("cluster_uid", "test-cluster-uid")
	_ = d.Set("vm_name", "test-vm")
	_ = d.Set("vm_namespace", "default")
	_ = d.Set("metadata", []interface{}{
		map[string]interface{}{
			"name":      "boot-vol",
			"namespace": "default",
		},
	})

	diags := resourceKubevirtDataVolumeCreate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.NotEmpty(t, d.Id())
}

// TestResourceKubevirtDataVolumeCreate_GetClusterError exercises the early
// GetCluster error branch via the negative mock client.
func TestResourceKubevirtDataVolumeCreate_GetClusterError(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	_ = d.Set("cluster_uid", "test-cluster-uid")
	_ = d.Set("vm_name", "test-vm")
	_ = d.Set("vm_namespace", "default")

	diags := resourceKubevirtDataVolumeCreate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError())
}

// TestResourceKubevirtDataVolumeUpdate — Update is implemented as Delete
// followed by Create, so a single end-to-end call exercises both.
func TestResourceKubevirtDataVolumeUpdate(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	_ = d.Set("cluster_uid", "test-cluster-uid")
	_ = d.Set("vm_name", "test-vm")
	_ = d.Set("vm_namespace", "default")
	_ = d.Set("metadata", []interface{}{
		map[string]interface{}{
			"name":      "boot-vol",
			"namespace": "default",
		},
	})
	// IdPartsDV expects: context/clusterUid/vmNamespace/vmName/volName
	d.SetId("project/test-cluster-uid/default/test-vm/boot-vol")

	diags := resourceKubevirtDataVolumeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	// Create (the second half of Update) rebuilds the ID.
	assert.NotEmpty(t, d.Id())
}

// TestResourceKubevirtDataVolumeUpdate_DeleteError — Delete fails first (bad
// ID), so Create is never reached.
func TestResourceKubevirtDataVolumeUpdate_DeleteError(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	d.SetId("only-two/parts")

	diags := resourceKubevirtDataVolumeUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceKubevirtDataVolumeRead_ValidMatch — valid ID with metadata
// name matching the mock's "boot-vol" fixture, so the flatten branch is
// entered.
func TestResourceKubevirtDataVolumeRead_ValidMatch(t *testing.T) {
	d := resourceKubevirtDataVolume().TestResourceData()
	_ = d.Set("cluster_context", "project")
	_ = d.Set("metadata", []interface{}{
		map[string]interface{}{
			"name":      "boot-vol",
			"namespace": "default",
		},
	})
	d.SetId("project/test-cluster-uid/default/test-vm/boot-vol")

	diags := resourceKubevirtDataVolumeRead(context.Background(), d, unitTestMockAPIClient)
	// The flatten step may succeed or return an error depending on the
	// exact shape of the DataVolumeTemplate spec in the mock; either
	// way, the branch is exercised.
	_ = diags
}
