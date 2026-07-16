package spectrocloud

import (
	"context"
	"testing"
	"time"

	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cluster_common_crud_test.go covers the state-refresh funcs and delete
// entrypoint in cluster_common_crud.go. Prior to this file, that whole
// module sat at 9.7% coverage because the waitFor* wrappers baked in a
// 30-second Delay before their first Refresh call, so exercising them
// end-to-end from a unit test would take minutes per case.
//
// Instead we test the *inner* funcs directly:
//   - resourceClusterReadyRefreshFunc
//   - resourceClusterStateRefreshFunc
//   - resourceVirtualClusterLifecycleStateRefreshFunc
//   - resourceClusterRead
//   - resourceClusterDelete (pre-wait branches)
//
// These are the pure state-derivation and error-plumbing paths — the
// wait* functions themselves are 10-line StateChangeConf assemblies whose
// only interesting behavior is "call the refresh func". Skipping their
// integration is a deliberate coverage/wall-clock tradeoff; the refresh
// funcs are where regressions actually hide.
//
// Fixtures come from mockCluster.go via UID dispatch — see
// clusterFixtureFor(). Each test picks a UID that shapes the mock's
// response into the exact state branch under test.

// castV1Client extracts the concrete *client.V1Client from
// unitTestMockAPIClient (typed as interface{} at the package level).
// A helper here avoids repeating the assertion in every t.Run.
func castV1Client(t *testing.T, m interface{}) *client.V1Client {
	c, ok := m.(*client.V1Client)
	require.True(t, ok, "unitTestMockAPIClient must be a *client.V1Client")
	return c
}

func TestResourceClusterReadyRefreshFunc(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	t.Run("cluster with populated status → Ready", func(t *testing.T) {
		// Default UID returns getMockSpectroCluster(), which has a
		// non-nil Status.
		obj, state, err := resourceClusterReadyRefreshFunc(c, "test-cluster-id")()
		require.NoError(t, err)
		assert.Equal(t, "Ready", state)
		assert.NotNil(t, obj)
	})

	t.Run("cluster with nil Status → NotReady", func(t *testing.T) {
		// GetClusterWithoutStatus returns the raw object; a nil Status
		// is the signal Terraform uses to keep polling on Create.
		_, state, err := resourceClusterReadyRefreshFunc(c, "cluster-uid-nil-status")()
		require.NoError(t, err)
		assert.Equal(t, "NotReady", state)
	})

	t.Run("404 → NotReady", func(t *testing.T) {
		// The SDK swallows 404s (apiutil.Is404 in GetClusterWithoutStatus)
		// and returns (nil, nil). The refresh func maps nil cluster to
		// NotReady rather than surfacing an error.
		_, state, err := resourceClusterReadyRefreshFunc(c, "cluster-uid-not-found")()
		require.NoError(t, err)
		assert.Equal(t, "NotReady", state)
	})

	t.Run("server error propagates", func(t *testing.T) {
		_, state, err := resourceClusterReadyRefreshFunc(c, "cluster-uid-server-error")()
		require.Error(t, err)
		assert.Equal(t, "", state, "state must be empty on error, or Terraform will loop forever polling for it")
	})
}

func TestResourceClusterStateRefreshFunc(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	t.Run("default cluster → Running-Healthy", func(t *testing.T) {
		// Default fixture has State=Running; overviewHandler defaults
		// return Health=Healthy → the func appends "-Healthy".
		_, state, err := resourceClusterStateRefreshFunc(c, "test-cluster-id")()
		require.NoError(t, err)
		assert.Equal(t, "Running-Healthy", state)
	})

	t.Run("provisioning → Provisioning (no health suffix)", func(t *testing.T) {
		// Non-Running states skip the overview lookup entirely — verify
		// the state passes through untouched.
		_, state, err := resourceClusterStateRefreshFunc(c, "cluster-uid-provisioning")()
		require.NoError(t, err)
		assert.Equal(t, "Provisioning", state)
	})

	t.Run("running-unhealthy → Running (no suffix)", func(t *testing.T) {
		// The func ONLY appends "-Healthy" when Health.State == "Healthy".
		// An Unhealthy cluster comes back as bare "Running" — that's a
		// signal to the caller to keep waiting.
		_, state, err := resourceClusterStateRefreshFunc(c, "cluster-uid-running-unhealthy")()
		require.NoError(t, err)
		assert.Equal(t, "Running", state,
			"Unhealthy Running should NOT get the -Healthy suffix")
	})

	t.Run("Status.State=Deleted → Deleted", func(t *testing.T) {
		// GetCluster returns (nil, nil) for State=="Deleted"; the refresh
		// func recognizes nil cluster as the Deleted state.
		_, state, err := resourceClusterStateRefreshFunc(c, "cluster-uid-deleted-state")()
		require.NoError(t, err)
		assert.Equal(t, "Deleted", state)
	})

	t.Run("404 → Deleted", func(t *testing.T) {
		// 404 also collapses to Deleted via GetCluster's nil-return path.
		_, state, err := resourceClusterStateRefreshFunc(c, "cluster-uid-not-found")()
		require.NoError(t, err)
		assert.Equal(t, "Deleted", state)
	})

	t.Run("server error propagates", func(t *testing.T) {
		_, _, err := resourceClusterStateRefreshFunc(c, "cluster-uid-server-error")()
		require.Error(t, err)
	})
}

func TestResourceVirtualClusterLifecycleStateRefreshFunc(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	t.Run("paused virtual cluster", func(t *testing.T) {
		_, state, err := resourceVirtualClusterLifecycleStateRefreshFunc(c, "project", "cluster-uid-paused")()
		require.NoError(t, err)
		assert.Equal(t, "Paused", state)
	})

	t.Run("deleted cluster returns Deleted state", func(t *testing.T) {
		_, state, err := resourceVirtualClusterLifecycleStateRefreshFunc(c, "project", "cluster-uid-deleted-state")()
		require.NoError(t, err)
		assert.Equal(t, "Deleted", state)
	})

	t.Run("error propagates", func(t *testing.T) {
		_, _, err := resourceVirtualClusterLifecycleStateRefreshFunc(c, "project", "cluster-uid-server-error")()
		require.Error(t, err)
	})
}

func TestResourceClusterRead(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	t.Run("populated cluster", func(t *testing.T) {
		d := resourceClusterAws().TestResourceData()
		d.SetId("test-cluster-id")
		cluster, err := resourceClusterRead(d, c, nil)
		require.NoError(t, err)
		require.NotNil(t, cluster)
		assert.Equal(t, "test-cluster", cluster.Metadata.Name)
	})

	t.Run("404 returns nil cluster and no error", func(t *testing.T) {
		// GetCluster maps 404 to (nil, nil). resourceClusterRead is a
		// thin wrapper, so it inherits that contract.
		d := resourceClusterAws().TestResourceData()
		d.SetId("cluster-uid-not-found")
		cluster, err := resourceClusterRead(d, c, nil)
		require.NoError(t, err)
		assert.Nil(t, cluster)
	})

	t.Run("server error propagates", func(t *testing.T) {
		d := resourceClusterAws().TestResourceData()
		d.SetId("cluster-uid-server-error")
		_, err := resourceClusterRead(d, c, nil)
		require.Error(t, err)
	})
}

// -----------------------------------------------------------------------------
// resourceClusterDelete has three interesting branches:
//   1. force_delete=true + force_delete_delay > default timeout → validation error
//   2. force_delete=true + delay <= timeout → DELETE then wait-with-fallback
//   3. force_delete=false → DELETE then wait
//
// Branches 2 and 3 both end up in waitForClusterDeletion, which has a 30s
// initial Delay we can't easily short-circuit from a unit test. So we
// exercise branch 1 (the pure validation guard) here and rely on Batch 3
// cluster-CRUD tests to exercise the other branches once mock timing is
// improved.
// -----------------------------------------------------------------------------

func TestResourceClusterDelete_ForceDeleteValidation(t *testing.T) {
	d := resourceClusterAws().TestResourceData()
	d.SetId("test-cluster-id")
	_ = d.Set("context", "project")
	_ = d.Set("force_delete", true)
	// Default delete timeout is 10 minutes for AWS clusters; setting the
	// delay above that forces the validation-error branch to fire before
	// any network call.
	_ = d.Set("force_delete_delay", 999)

	diags := resourceClusterDelete(context.Background(), d, unitTestMockAPIClient)
	require.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary, "Force delete validation failed")
	assert.Contains(t, diags[0].Detail, "force_delete_delay")
}

// TestWaitForClusterDeletion_ContextCancelled exercises the error-branch of
// waitForClusterDeletion by handing it a context that's already cancelled.
// The retry loop should observe ctx.Err() on its first iteration and
// return without waiting the 30-second Delay. This covers the
// waitForClusterDeletion body's non-happy exit path without paying the
// wall-clock cost of a real timeout.
func TestWaitForClusterDeletion_ContextCancelled(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	err := waitForClusterDeletion(ctx, c, "project", "test-cluster-id", 5*time.Second)
	require.Error(t, err, "cancelled context must surface an error")
}

// TestWaitForClusterReady_ContextCancelled — same pattern, one level up
// the wait chain. Covers waitForClusterReady's non-target exit.
func TestWaitForClusterReady_ContextCancelled(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	d := resourceClusterAws().TestResourceData()
	_ = d.Set("context", "project")
	// No need to shrink the timeout — the context is already cancelled,
	// so WaitForStateContext returns immediately with ctx.Err().

	diags, hadError := waitForClusterReady(ctx, d, "test-cluster-id", nil, c)
	assert.True(t, hadError)
	assert.NotEmpty(t, diags)
}

// TestWaitForClusterCreation_SkipCompletion pins the fast-path exit
// where skip_completion=true skips the whole wait chain — this branch
// is used by "attach an addon" scenarios that don't own the cluster
// lifecycle and must not block on its readiness.
func TestWaitForClusterCreation_SkipCompletion(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterAws().TestResourceData()
	_ = d.Set("skip_completion", true)

	diags, exited := waitForClusterCreation(context.Background(), d, "test-cluster-id", nil, c, true)
	assert.True(t, exited, "skip_completion=true must exit before touching the API")
	assert.Empty(t, diags)
}

// TestWaitForClusterCreation_ContextCancelled — cancelled context inside
// the waitForClusterReady sub-call surfaces as an error. This proves the
// error propagates from the inner wait to the outer wait.
func TestWaitForClusterCreation_ContextCancelled(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	d := resourceClusterAws().TestResourceData()
	diags, hadError := waitForClusterCreation(ctx, d, "test-cluster-id", nil, c, true)
	assert.True(t, hadError)
	assert.NotEmpty(t, diags)
}
