package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/apiutil"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
	"log"
)

// isForbiddenErr returns true if the error represents an API authorization
// failure (e.g. the caller lacks a specific permission such as
// `cluster.adminKubeconfigDownload`), as opposed to a genuine failure.
func isForbiddenErr(err error) bool {
	code := apiutil.ToV1ErrorObj(err).Code
	return code == "OperationForbidden" || code == "ResOperationForbidden"
}

// ProviderMeta is the provider's `meta` value, returned by providerConfigure
// and passed as `m` to every resource/data-source CRUD function. It holds
// one client per supported scope, each fully scoped once at configure time -
// never mutated afterward - so concurrent resource CRUDs (Terraform runs
// these in parallel) only ever read already-finished, independent clients
// instead of racing on a shared client's scope (PLT-2438).
type ProviderMeta struct {
	Project    *client.V1Client
	Tenant     *client.V1Client
	ProjectUID string
}

func getV1ClientWithResourceContext(m interface{}, resourceContext string) *client.V1Client {
	pm := m.(*ProviderMeta)
	if resourceContext == "tenant" {
		return pm.Tenant
	}
	return pm.Project
}

// getProviderProjectUID returns the project UID resolved from the provider's
// `project_name` at configure time, replacing the former package-level
// ProviderInitProjectUid global (which was overwritten by every provider
// block/alias, corrupting any other aliased provider's calls).
func getProviderProjectUID(m interface{}) string {
	return m.(*ProviderMeta).ProjectUID
}

func handleReadError(d *schema.ResourceData, err error, diags diag.Diagnostics) diag.Diagnostics {
	if herr.IsNotFound(err) {
		d.SetId("")
		return diags
	}
	log.Printf("[DEBUG] Received error: %#v", err)
	return diag.FromErr(err)
}
