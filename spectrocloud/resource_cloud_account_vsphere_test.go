package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
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
			VcenterServer: types.Ptr("vcenter.example.com"),
			Username:      types.Ptr("testuser"),
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
