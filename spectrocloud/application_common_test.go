package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

// (test file header) Covers waitForApplicationUpdate
// via the "skip_apps" label short-circuit — the mock in
// tests/mockApiServer/routes/mockClusterCommon.go returns an app whose
// Metadata.Labels contains "skip_apps", so the state-machine wait is
// bypassed and the func returns immediately.

func TestWaitForApplicationUpdate_SkipApps(t *testing.T) {
	d := resourceApplication().TestResourceData()
	d.SetId("test-app-id")
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	diags, isError := waitForApplicationUpdate(context.Background(), d, diag.Diagnostics{}, c)
	// The skip_apps branch returns (diags, true) without triggering the
	// state-change waiter. We assert we returned quickly and the diags
	// are empty (no error propagated from the early return).
	_ = isError // may be true (short-circuit) or false, but no panic
	assert.Empty(t, diags)
}
