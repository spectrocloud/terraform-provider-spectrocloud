package spectrocloud

import (
	"fmt"
	"testing"

	"github.com/spectrocloud/hapi/client"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func TestNameSpacesRPSScenario(t *testing.T) {
	if !IsIntegrationTestEnvSet(baseConfig) {
		t.Skip("Skipping integration test env variable not set")
	}
	cases := []Retry{
		{50, 1, 429},
	}

	for _, c := range cases {
		h := client.New("api.dev.spectrocloud.com", "nikolay@spectrocloud.com", "", "", "mB9NKY0bBRlYy4eP1wteirbxHQ0b5ypL", false, c.retries)
		uid, err := h.GetProjectUID("Default")
		if err != nil {
			t.Fail()
		}
		projectH := client.New("api.dev.spectrocloud.com", "nikolay@spectrocloud.com", "", uid, "mB9NKY0bBRlYy4eP1wteirbxHQ0b5ypL", false, c.retries)
		GetNamespaces1Test(t, projectH, c)
	}
}

// 1. Normal case where rps is just within the limit. 5 rps or 50 with burst. Expected result: no retries, no errors.
func GetNamespaces1Test(t *testing.T, h *client.V1Client, retry Retry) {
	client, err := h.GetClusterClient()
	if err != nil {
		t.Fail()
	}

	cluster, err := h.GetClusterByName("eks-dev-nik-4", "project")
	if err != nil && cluster == nil {
		t.Fail()
	}

	params := clusterC.NewV1SpectroClustersUIDConfigNamespacesGetParamsWithContext(h.Ctx).WithUID(cluster.Metadata.UID)

	// 2. Many requests but retry works. For example for 100 rps, 1 retry_attempt yeilds no erros.
	// (default timeout for retry is starting at 2 seconds, and exponentially increasing with jitter)
	// jitter := time.Duration(rand.Int63n(int64(sleep)))
	// sleep = (2 * sleep) + jitter/2 //exponential sleep with jitter. 2,

	// 3. Too many requests that retry stops working. 1 retry_attempt but we invoke just enough requests concurrently to cause some number(20% ,33%) of them to exist with 429.
	// But also check that request indeed was retried.
	ch := make(chan int)
	done := make(chan bool)

	method, in := prepareClusterMethod(client, params, "V1SpectroClustersUIDConfigNamespacesGet")
	go produceResults(retry, method, in, ch, done)

	stat := consumeResults(t, retry, ch, done)
	fmt.Printf("\nDone: %d, %d, %d, %d.\n", stat.CODE_MINUS_ONE, stat.CODE_NORMAL, stat.CODE_EXPECTED, stat.CODE_INTERNAL_ERROR)
}
