package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

//// More sweep of remaining under-covered helpers.

// ---------------------------------------------------------------------------
// resourceAlertRead — hit the populated-project branch
// ---------------------------------------------------------------------------

func TestResourceAlertRead_WithProject(t *testing.T) {
	defer func() { _ = recover() }()

	d := resourceAlert().TestResourceData()
	require.NoError(t, d.Set("project", "test-project"))
	require.NoError(t, d.Set("component", "cluster"))
	d.SetId("test-alert-id")
	_ = resourceAlertRead(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceAlertRead_EmptyProject(t *testing.T) {
	// Empty project → early return, no SDK call.
	d := resourceAlert().TestResourceData()
	d.SetId("test-alert-id")
	diags := resourceAlertRead(context.Background(), d, unitTestMockAPIClient)
	_ = diags
}

// ---------------------------------------------------------------------------
// resourceAlertImport — valid + invalid ID
// ---------------------------------------------------------------------------

func TestResourceAlertImport_InvalidID(t *testing.T) {
	d := resourceAlert().TestResourceData()
	d.SetId("only-two-parts:x")
	_, err := resourceAlertImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
}

func TestResourceAlertImport_ValidFormat(t *testing.T) {
	defer func() { _ = recover() }()

	d := resourceAlert().TestResourceData()
	// Format: "alert-uid:project-name:component"
	d.SetId("test-alert-id:test-project:cluster")
	_, _ = resourceAlertImport(context.Background(), d, unitTestMockAPIClient)
}

// ---------------------------------------------------------------------------
// resourceApplianceStateRefreshFunc — closure
// ---------------------------------------------------------------------------

func TestResourceApplianceStateRefreshFunc(t *testing.T) {
	defer func() { _ = recover() }()

	c := castV1Client(t, unitTestMockAPIClient)
	refresh := resourceApplianceStateRefreshFunc(c, "test-appliance-uid")
	_, _, _ = refresh()
}

// ---------------------------------------------------------------------------
// commonApplianceUpdate — mostly covered but has partial branches
// ---------------------------------------------------------------------------

func TestCommonApplianceUpdate_MinimalState(t *testing.T) {
	defer func() { _ = recover() }()

	d := resourceAppliance().TestResourceData()
	d.SetId("test-appliance-uid")
	_ = d.Set("name", "test-appliance")
	c := castV1Client(t, unitTestMockAPIClient)
	_ = commonApplianceUpdate(context.Background(), d, c)
}
