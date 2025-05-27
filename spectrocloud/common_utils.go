package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
	"log"
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

func handleReadError(d *schema.ResourceData, err error, diags diag.Diagnostics) diag.Diagnostics {
	if herr.IsNotFound(err) {
		d.SetId("")
		return diags
	}
	log.Printf("[DEBUG] Received error: %#v", err)
	return diag.FromErr(err)
}
