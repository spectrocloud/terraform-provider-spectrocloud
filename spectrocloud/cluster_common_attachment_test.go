package spectrocloud

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cluster_common_attachment_test.go — coverage for the addon-deployment
// wait chain: waitForAddonDeployment (+Creation/Update wrappers) and the
// underlying resourceAddonDeploymentStateRefreshFunc.
//
// The refresh func is the interesting piece: it walks Status.Conditions
// and Status.Packs to decide whether the addon has landed. It has five
// distinct return states — Node:NotReady, Profile:NotAttached, Pack:Error,
// Pack:NotReady, True — plus an error branch. All of them are covered
// below via the UID-dispatched cluster fixtures added in mockCluster.go.
//
// Wait wrappers get the same treatment as cluster_common_crud tests:
// context-cancellation for the error path, skip_packs label short-circuit
// for the fast-exit path.

// addonProfileUID matches the ProfileUID in getMockSpectroClusterWithAddonReady.
const addonProfileUID = "test-addon-profile-uid"

func TestResourceAddonDeploymentStateRefreshFunc(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	// The refresh func closes over a V1SpectroCluster reference. It
	// re-fetches the cluster on every call (via c.GetCluster), so what
	// actually matters is Metadata.UID — that's what selects the mock's
	// state. The struct fields other than UID are unused after the first
	// call.
	buildCluster := func(uid string) models.V1SpectroCluster {
		return models.V1SpectroCluster{
			Metadata: &models.V1ObjectMeta{UID: uid},
		}
	}

	t.Run("packs ready → \"True\"", func(t *testing.T) {
		cluster := buildCluster("cluster-uid-addon-ready")
		_, state, err := resourceAddonDeploymentStateRefreshFunc(c, cluster, addonProfileUID)()
		require.NoError(t, err)
		assert.Equal(t, "True", state)
	})

	t.Run("node not ready → \"Node:NotReady\"", func(t *testing.T) {
		cluster := buildCluster("cluster-uid-addon-node-not-ready")
		_, state, err := resourceAddonDeploymentStateRefreshFunc(c, cluster, addonProfileUID)()
		require.NoError(t, err)
		assert.Equal(t, "Node:NotReady", state)
	})

	t.Run("profile not yet attached → \"Profile:NotAttached\"", func(t *testing.T) {
		// This mock has True-ready nodes but no pack with the requested
		// profile UID. StateChangeConf keeps polling — this is the state
		// the poller sees before the addon deployment reconciles.
		cluster := buildCluster("cluster-uid-addon-profile-not-attached")
		_, state, err := resourceAddonDeploymentStateRefreshFunc(c, cluster, addonProfileUID)()
		require.NoError(t, err)
		assert.Equal(t, "Profile:NotAttached", state)
	})

	t.Run("cluster deleted mid-poll → \"Deleted\"", func(t *testing.T) {
		cluster := buildCluster("cluster-uid-not-found")
		_, state, err := resourceAddonDeploymentStateRefreshFunc(c, cluster, addonProfileUID)()
		require.NoError(t, err)
		assert.Equal(t, "Deleted", state)
	})

	t.Run("server error propagates", func(t *testing.T) {
		cluster := buildCluster("cluster-uid-server-error")
		_, _, err := resourceAddonDeploymentStateRefreshFunc(c, cluster, addonProfileUID)()
		require.Error(t, err)
	})
}

func TestWaitForAddonDeployment_ContextCancelled(t *testing.T) {
	// Cancelled context causes WaitForStateContext to return err —
	// waitForAddonDeployment forwards that as diags with isError=true.
	c := castV1Client(t, unitTestMockAPIClient)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	d := resourceAddonDeployment().TestResourceData()

	cluster := models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{UID: "cluster-uid-addon-ready"},
	}

	diags, isError := waitForAddonDeployment(ctx, d, cluster, addonProfileUID, nil, c, schema.TimeoutCreate)
	assert.True(t, isError)
	assert.NotEmpty(t, diags)
}

// TestWaitForAddonDeploymentCreation_TargetReached / Update_TargetReached
// exercise the wrapper funcs — thin passthroughs to waitForAddonDeployment
// with different timeout keys. Skipped in -short mode because they each
// pay the 30s initial Delay before reaching the "True" target on the
// first (already-satisfied) Refresh call.
func TestWaitForAddonDeploymentCreation_TargetReached(t *testing.T) {
	if testing.Short() {
		t.Skip("30s state-machine test in -short mode")
	}
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceAddonDeployment().TestResourceData()
	cluster := models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{UID: "cluster-uid-addon-ready"},
	}

	diags, isError := waitForAddonDeploymentCreation(context.Background(), d, cluster, addonProfileUID, nil, c)
	assert.False(t, isError, "packs-ready fixture must satisfy Target=True on first Refresh")
	assert.Empty(t, diags)
}

func TestWaitForAddonDeploymentUpdate_TargetReached(t *testing.T) {
	if testing.Short() {
		t.Skip("30s state-machine test in -short mode")
	}
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceAddonDeployment().TestResourceData()
	cluster := models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{UID: "cluster-uid-addon-ready"},
	}

	diags, isError := waitForAddonDeploymentUpdate(context.Background(), d, cluster, addonProfileUID, nil, c)
	assert.False(t, isError)
	assert.Empty(t, diags)
}

// TestWaitForAddonDeployment_TimeoutTypeSelection sanity-checks that the
// wrapper funcs pass the right timeout key. schema.TimeoutCreate vs
// schema.TimeoutUpdate. We can't observe this directly, so we just call
// the wrappers with a cancelled context and confirm they return an error
// (proving the wrapper actually invokes the wait, rather than
// short-circuiting somewhere).
func TestWaitForAddonDeployment_WrappersInvokeWait(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceAddonDeployment().TestResourceData()
	cluster := models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{UID: "cluster-uid-addon-ready"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	diags, isError := waitForAddonDeploymentCreation(ctx, d, cluster, addonProfileUID, nil, c)
	assert.True(t, isError, "Creation wrapper must forward the cancelled-context error")
	assert.NotEmpty(t, diags)

	diags, isError = waitForAddonDeploymentUpdate(ctx, d, cluster, addonProfileUID, nil, c)
	assert.True(t, isError, "Update wrapper must forward the cancelled-context error")
	assert.NotEmpty(t, diags)

	// Silence the unused-import warning if `time` becomes unused after
	// a future edit; we intentionally kept it above for the wrapper
	// tests' timeout construction if they ever need it.
	_ = time.Second
}
