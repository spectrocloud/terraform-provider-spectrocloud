package spectrocloud

import (
	"fmt"
	"testing"

	userC "github.com/spectrocloud/hapi/user/client/v1"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func Test1Scenario(t *testing.T) {
	if !IsIntegrationTestEnvSet(baseConfig) {
		t.Skip("Skipping integration test env variable not set")
	}
	cases := []Retry{
		{190, 3, 429},
	}

	for _, c := range cases {
		h := client.New(
			client.WithHubbleURI(baseConfig.hubbleHost),
			client.WithAPIKey(baseConfig.apikey),
			client.WithRetries(c.retries))
		uid, err := h.GetProjectUID(baseConfig.project)
		if err != nil {
			t.Fail()
		}
		client.WithProjectUID(uid)(h)
		GetProjects1Test(t, h, c)
	}
}

// 1. Normal case where rps is just within the limit. 5 rps or 50 with burst. Expected result: no retries, no errors.
func GetProjects1Test(t *testing.T, h *client.V1Client, retry Retry) {
	userClient := h.GetUserClient()

	limit := int64(0)
	params := userC.NewV1ProjectsListParams().WithLimit(&limit)

	// 2. Many requests but retry works. For example for 100 rps, 1 retry_attempt yeilds no erros.
	// (default timeout for retry is starting at 2 seconds, and exponentially increasing with jitter)
	// jitter := time.Duration(rand.Int63n(int64(sleep)))
	// sleep = (2 * sleep) + jitter/2 //exponential sleep with jitter. 2,

	// 3. Too many requests that retry stops working. 1 retry_attempt but we invoke just enough requests concurrently to cause some number(20% ,33%) of them to exist with 429.
	// But also check that request indeed was retried.
	ch := make(chan int)
	done := make(chan bool)

	method, in := prepareUserMethod(userClient, params, "V1ProjectsList")
	go produceResults(retry, method, in, ch, done)

	stat := consumeResults(t, retry, ch, done)
	fmt.Printf("\nDone: %d, %d, %d, %d.\n", stat.CODE_MINUS_ONE, stat.CODE_NORMAL, stat.CODE_EXPECTED, stat.CODE_INTERNAL_ERROR)
}
