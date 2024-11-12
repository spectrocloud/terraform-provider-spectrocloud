package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func TestToVsphereAccount(t *testing.T) {
	rd := resourceCloudAccountVsphere().TestResourceData()
	rd.Set("name", "vsphere_unit_test_acc")
	rd.Set("vsphere_vcenter", "vcenter.example.com")
	rd.Set("vsphere_username", "testuser")
	rd.Set("vsphere_password", "testpass")
	rd.Set("vsphere_ignore_insecure_error", false)
	rd.Set("private_cloud_gateway_id", "12345")
	acc := toVsphereAccount(rd)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("vsphere_vcenter"), *acc.Spec.VcenterServer)
	assert.Equal(t, rd.Get("vsphere_username"), *acc.Spec.Username)
	assert.Equal(t, rd.Get("vsphere_password"), *acc.Spec.Password)
	assert.Equal(t, rd.Get("vsphere_ignore_insecure_error"), acc.Spec.Insecure)
	assert.Equal(t, rd.Get("private_cloud_gateway_id"), acc.Metadata.Annotations[OverlordUID])
	assert.Equal(t, rd.Id(), acc.Metadata.UID)
}

func TestToVsphereAccountIgnoreInsecureError(t *testing.T) {
	rd := resourceCloudAccountVsphere().TestResourceData()
	rd.Set("name", "vsphere_unit_test_acc")
	rd.Set("context", "tenant")
	rd.Set("vsphere_vcenter", "vcenter.example.com")
	rd.Set("vsphere_username", "testuser")
	rd.Set("vsphere_password", "testpass")
	rd.Set("vsphere_ignore_insecure_error", true)
	rd.Set("private_cloud_gateway_id", "67890")
	acc := toVsphereAccount(rd)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("vsphere_vcenter"), *acc.Spec.VcenterServer)
	assert.Equal(t, rd.Get("vsphere_username"), *acc.Spec.Username)
	assert.Equal(t, rd.Get("vsphere_password"), *acc.Spec.Password)
	assert.Equal(t, rd.Get("vsphere_ignore_insecure_error"), acc.Spec.Insecure)
	assert.Equal(t, rd.Get("private_cloud_gateway_id"), acc.Metadata.Annotations[OverlordUID])
	assert.Equal(t, rd.Id(), acc.Metadata.UID)
}

func TestFlattenVsphereCloudAccount(t *testing.T) {
	rd := resourceCloudAccountVsphere().TestResourceData()
	account := &models.V1VsphereAccount{
		Metadata: &models.V1ObjectMeta{
			Name:        "test_account",
			Annotations: map[string]string{OverlordUID: "12345"},
			UID:         "abcdef",
		},
		Spec: &models.V1VsphereCloudAccount{
			VcenterServer: ptr.To("vcenter.example.com"),
			Username:      ptr.To("testuser"),
			Insecure:      true,
		},
	}

	diags, hasError := flattenVsphereCloudAccount(rd, account)

	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "test_account", rd.Get("name"))
	assert.Equal(t, "12345", rd.Get("private_cloud_gateway_id"))
	assert.Equal(t, "vcenter.example.com", rd.Get("vsphere_vcenter"))
	assert.Equal(t, "testuser", rd.Get("vsphere_username"))
	assert.Equal(t, true, rd.Get("vsphere_ignore_insecure_error"))
}

func prepareResourceCloudAccountVsphere() *schema.ResourceData {
	d := resourceCloudAccountVsphere().TestResourceData()
	d.SetId("test-vsphere-account-id-1")
	_ = d.Set("name", "test-vsphere-account-1")
	_ = d.Set("context", "project")
	_ = d.Set("private_cloud_gateway_id", "pcg-id")
	_ = d.Set("vsphere_vcenter", "test-vcenter")
	_ = d.Set("vsphere_username", "test-uname")
	_ = d.Set("vsphere_password", "test-pwd")
	_ = d.Set("vsphere_ignore_insecure_error", false)
	return d
}

func TestResourceCloudAccountVsphereCreate(t *testing.T) {
	d := prepareResourceCloudAccountVsphere()
	ctx := context.Background()
	diags := resourceCloudAccountVsphereCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-vsphere-account-id-1", d.Id())
}

func TestResourceCloudAccountVsphereRead(t *testing.T) {
	d := prepareResourceCloudAccountVsphere()
	ctx := context.Background()
	diags := resourceCloudAccountVsphereRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-vsphere-account-id-1", d.Id())
}

func TestResourceCloudAccountVsphereUpdate(t *testing.T) {
	d := prepareResourceCloudAccountVsphere()
	ctx := context.Background()
	diags := resourceCloudAccountVsphereUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-vsphere-account-id-1", d.Id())
}

func TestResourceCloudAccountVsphereDelete(t *testing.T) {
	d := prepareResourceCloudAccountVsphere()
	ctx := context.Background()
	diags := resourceCloudAccountVsphereDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}
