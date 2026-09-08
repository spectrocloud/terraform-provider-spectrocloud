package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// TestResourceClusterAwsRead_AdminKubeConfigForbidden covers PLT-2355: a
// caller without the cluster.adminKubeconfigDownload permission gets a 403
// from GetClusterAdminKubeConfig on every plan/apply. Read must tolerate
// that specific failure (leaving admin_kube_config unset) instead of
// aborting the whole read, since the permission is unrelated to the
// caller's ability to manage the cluster itself.
func TestResourceClusterAwsRead_AdminKubeConfigForbidden(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId(routes.AdminKubeConfigForbiddenUID)

	diags := resourceClusterAwsRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Empty(t, d.Get("admin_kube_config"))
}
