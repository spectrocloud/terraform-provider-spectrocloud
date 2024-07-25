package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/client"
)

func setResourceContext(c *client.V1Client, resourceContext string) {
	switch resourceContext {
	case "tenant":
		client.WithScopeTenant()(c)
	default:
		if ProviderInitProjectUid != "" {
			client.WithScopeProject(ProviderInitProjectUid)(c)
		}
	}
}

// setResourceContext(c, ProjectContext)
