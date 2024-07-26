package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/client"
)

func getV1ClientWithResourceContext(m interface{}, resourceContext string) *client.V1Client {
	c := m.(*client.V1Client)
	switch resourceContext {
	case "project":
		if ProviderInitProjectUid != "" {
			client.WithScopeProject(ProviderInitProjectUid)(c)
		}
		return c
	case "tenant":
		client.WithScopeTenant()(c)
		return c
	default:
		if ProviderInitProjectUid != "" {
			client.WithScopeProject(ProviderInitProjectUid)(c)
		}
		return c
	}
}

// setResourceContext(c, ProjectContext)
