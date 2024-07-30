package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/client"
)

func TestNameSpacesRPSScenario(t *testing.T) {
	if !IsIntegrationTestEnvSet(baseConfig) {
		t.Skip("Skipping integration test env variable not set")
	}
	cases := []Retry{
		{50, 1, 429},
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
