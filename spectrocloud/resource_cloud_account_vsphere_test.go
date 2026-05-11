package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestToVsphereAccount_TableDriven(t *testing.T) {
	tests := []struct {
		name   string
		rdSet  map[string]interface{}
		verify func(t *testing.T, rd *schema.ResourceData, acc *models.V1VsphereAccount)
	}{
		{
			name: "base account",
			rdSet: map[string]interface{}{
				"name": "vsphere_unit_test_acc", "vsphere_vcenter": "vcenter.example.com",
				"vsphere_username": "testuser", "vsphere_password": "testpass",
				"vsphere_ignore_insecure_error": false, "private_cloud_gateway_id": "12345",
			},
			verify: func(t *testing.T, rd *schema.ResourceData, acc *models.V1VsphereAccount) {
				assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
				assert.Equal(t, rd.Get("vsphere_vcenter"), *acc.Spec.VcenterServer)
				assert.Equal(t, rd.Get("vsphere_username"), *acc.Spec.Username)
				assert.Equal(t, rd.Get("vsphere_password"), *acc.Spec.Password)
				assert.Equal(t, rd.Get("vsphere_ignore_insecure_error"), acc.Spec.Insecure)
				assert.Equal(t, rd.Get("private_cloud_gateway_id"), acc.Metadata.Annotations[OverlordUID])
				assert.Equal(t, rd.Id(), acc.Metadata.UID)
			},
		},
		{
			name: "ignore insecure error with tenant context",
			rdSet: map[string]interface{}{
				"name": "vsphere_unit_test_acc", "context": "tenant",
				"vsphere_vcenter": "vcenter.example.com", "vsphere_username": "testuser",
				"vsphere_password": "testpass", "vsphere_ignore_insecure_error": true,
				"private_cloud_gateway_id": "67890",
			},
			verify: func(t *testing.T, rd *schema.ResourceData, acc *models.V1VsphereAccount) {
				assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
				assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
				assert.Equal(t, rd.Get("vsphere_vcenter"), *acc.Spec.VcenterServer)
				assert.Equal(t, rd.Get("vsphere_ignore_insecure_error"), acc.Spec.Insecure)
				assert.Equal(t, rd.Get("private_cloud_gateway_id"), acc.Metadata.Annotations[OverlordUID])
				assert.Equal(t, rd.Id(), acc.Metadata.UID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := resourceCloudAccountVsphere().TestResourceData()
			for k, v := range tt.rdSet {
				rd.Set(k, v)
			}
			acc := toVsphereAccount(rd)
			tt.verify(t, rd, acc)
		})
	}
}

func TestFlattenVsphereCloudAccount_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		account *models.V1VsphereAccount
		expect  map[string]interface{}
	}{
		{
			name: "base account",
			account: &models.V1VsphereAccount{
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
			},
			expect: map[string]interface{}{
				"name": "test_account", "private_cloud_gateway_id": "12345",
				"vsphere_vcenter": "vcenter.example.com", "vsphere_username": "testuser",
				"vsphere_ignore_insecure_error": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := resourceCloudAccountVsphere().TestResourceData()
			diags, hasError := flattenVsphereCloudAccount(rd, tt.account)
			assert.Nil(t, diags)
			assert.False(t, hasError)
			for k, want := range tt.expect {
				assert.Equal(t, want, rd.Get(k), "field %s", k)
			}
		})
	}
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

func TestResourceCloudAccountVsphereCRUD_TableDriven(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		op   func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	}{
		{name: "Create", op: resourceCloudAccountVsphereCreate},
		{name: "Read", op: resourceCloudAccountVsphereRead},
		{name: "Update", op: resourceCloudAccountVsphereUpdate},
		{name: "Delete", op: resourceCloudAccountVsphereDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := prepareResourceCloudAccountVsphere()
			diags := tt.op(ctx, d, unitTestMockAPIClient)
			assert.Len(t, diags, 0)
			if tt.name != "Delete" {
				assert.Equal(t, "test-vsphere-account-id-1", d.Id())
			}
		})
	}
}
