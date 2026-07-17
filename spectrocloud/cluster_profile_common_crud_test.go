package spectrocloud

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cluster_profile_common_crud.go was the last file in the top-level
// spectrocloud package sitting at 0% coverage. It has two funcs, both
// wrapped around the StateChangeConf retry loop:
//
//   waitForProfileDownload — polls until SpcApply.CanBeApplied flips true.
//   resourceClusterProfileStateRefreshFunc — reads the cluster and
//     stringifies canBeApplied for the poller.
//
// Same testing strategy as cluster_common_crud: call the refresh func
// directly against the mock (fast, deterministic), and drive
// waitForProfileDownload's error path via context cancellation instead
// of paying its 30-second Delay.

func TestResourceClusterProfileStateRefreshFunc(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	t.Run("canBeApplied=true → \"true\"", func(t *testing.T) {
		// Default cluster fixture has SpcApply.CanBeApplied=true (see
		// getMockSpectroCluster in tests/mockApiServer/routes/mockCluster.go).
		_, state, err := resourceClusterProfileStateRefreshFunc(c, "test-cluster-id")()
		require.NoError(t, err)
		assert.Equal(t, "true", state,
			"canBeApplied=true must serialize to lowercase \"true\" (matches waitForProfileDownload's Target)")
	})

	t.Run("canBeApplied=false → \"false\"", func(t *testing.T) {
		_, state, err := resourceClusterProfileStateRefreshFunc(c, "cluster-uid-spcapply-false")()
		require.NoError(t, err)
		assert.Equal(t, "false", state)
	})

	t.Run("cluster deleted mid-poll → \"Cluster deleted\"", func(t *testing.T) {
		// GetCluster returns (nil, nil) when the cluster is gone. The
		// refresh func maps this to a sentinel state string that
		// StateChangeConf sees as neither Pending nor Target — the loop
		// exits with an "unexpected state" error, which is what the
		// caller wants: don't hang forever if the profile's cluster
		// went away.
		_, state, err := resourceClusterProfileStateRefreshFunc(c, "cluster-uid-not-found")()
		require.NoError(t, err)
		assert.Equal(t, "Cluster deleted", state)
	})

	t.Run("server error propagates", func(t *testing.T) {
		_, _, err := resourceClusterProfileStateRefreshFunc(c, "cluster-uid-server-error")()
		require.Error(t, err)
	})
}

func TestWaitForProfileDownload_ContextCancelled(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	// A cancelled context makes WaitForStateContext return immediately;
	// we don't pay the 30s initial Delay. This exercises the error
	// return of waitForProfileDownload without the wait.
	err := waitForProfileDownload(ctx, c, "project", "test-cluster-id", 5*time.Second)
	require.Error(t, err)
}

func TestWaitForProfileDownload_TargetReached(t *testing.T) {
	// The default cluster fixture already reports canBeApplied=true, so
	// the very first Refresh call hits the Target state. StateChangeConf
	// still waits its 30s Delay before that first call — meaning this
	// test takes ~30s. Marked short-skip so `go test -short` stays fast;
	// the CI/coverage run picks it up.
	if testing.Short() {
		t.Skip("skipping 30s state-machine test in -short mode")
	}
	c := castV1Client(t, unitTestMockAPIClient)
	err := waitForProfileDownload(context.Background(), c, "project", "test-cluster-id", 2*time.Minute)
	assert.NoError(t, err, "canBeApplied=true from the mock's first response must satisfy Target")
}
