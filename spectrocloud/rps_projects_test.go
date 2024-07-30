package spectrocloud

import (
	"testing"

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
			client.WithPaletteURI(baseConfig.hubbleHost),
			client.WithAPIKey(baseConfig.apikey),
			client.WithRetries(c.retries))
		uid, err := h.GetProjectUID(baseConfig.project)
		if err != nil {
			t.Fail()
		}
		client.WithScopeProject(uid)(h)
	}
}
