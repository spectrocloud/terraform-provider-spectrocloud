package test

import (
	"fmt"
	"github.com/spectrocloud/hapi/apiutil/transport"
	userC "github.com/spectrocloud/hapi/user/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"testing"
)

type Retry struct {
	runs          int
	retries       int
	expected_code int
}

type ResultStat struct {
	CODE_MINUS_ONE      int
	CODE_NORMAL         int
	CODE_EXPECTED       int
	CODE_INTERNAL_ERROR int
}

func Test1Scenario(t *testing.T) {
	cases := []Retry{
		{290, 3, 429},
	}

	for _, c := range cases {
		h := client.New("api.dev.spectrocloud.com", "nikolay@spectrocloud.com", "", "Default", "QR5aRhZe0XZjP2bvLDEcToC0xBBqgmjS", false, c.retries)
		GetProjects1Test(t, h, c)
	}
}

// 1. Normal case where rps is just within the limit. 5 rps or 50 with burst. Expected result: no retries, no errors.
func GetProjects1Test(t *testing.T, h *client.V1Client, retry Retry) {
	userClient, err := h.GetUserClient()

	if err != nil {
		t.Fail()
	}

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

	go produceResults(retry, userClient, params, ch, done)

	stat := consumeResults(t, retry, ch, done)
	fmt.Printf("\nDone: %d, %d, %d, %d.\n", stat.CODE_MINUS_ONE, stat.CODE_NORMAL, stat.CODE_EXPECTED, stat.CODE_INTERNAL_ERROR)
}

func consumeResults(t *testing.T, retry Retry, ch chan int, done chan bool) ResultStat {
	stat := ResultStat{
		CODE_MINUS_ONE:      0,
		CODE_NORMAL:         0,
		CODE_EXPECTED:       0,
		CODE_INTERNAL_ERROR: 0,
	}

	for i := 0; i < retry.runs; i++ {
		v := <-ch
		switch v {
		case -1:
			stat.CODE_MINUS_ONE++
			break
		case retry.expected_code:
			stat.CODE_EXPECTED++
			break
		case 200:
			stat.CODE_NORMAL++
			break
		case 500:
			stat.CODE_INTERNAL_ERROR++
			break
		default:
			t.Fail()
		}
	}
	<-done
	return stat
}

func produceResults(retry Retry, userClient userC.ClientService, params *userC.V1ProjectsListParams, ch chan int, done chan bool) {
	for i := 0; i < retry.runs; i++ {
		go func(chnl chan int) {
			_, err := userClient.V1ProjectsList(params)
			if err != nil {
				if _, ok := err.(*transport.TcpError); ok {
					chnl <- -1
					return
				}
				if _, ok := err.(*transport.TransportError); ok && err.(*transport.TransportError).HttpCode == retry.expected_code {
					chnl <- retry.expected_code
					return
				} else {
					chnl <- 500
					return
				}
			} else {
				chnl <- 200
				return
			}
		}(ch)
	}
	done <- true
}
