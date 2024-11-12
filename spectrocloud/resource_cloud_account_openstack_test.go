package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

// Test for the `toOpenStackAccount` function
func TestToOpenStackAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1OpenStackAccount
	}{
		{
			name: "Valid Data",
			input: map[string]interface{}{
				"name":                     "openstack-account",
				"private_cloud_gateway_id": "private-cloud-gateway-id",
				"ca_certificate":           "ca-cert",
				"default_domain":           "default-domain",
				"default_project":          "default-project",
				"identity_endpoint":        "http://identity.endpoint",
				"openstack_allow_insecure": true,
				"parent_region":            "parent-region",
				"openstack_password":       "password",
				"openstack_username":       "username",
			},
			expected: &models.V1OpenStackAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "openstack-account",
					Annotations: map[string]string{OverlordUID: "private-cloud-gateway-id"},
					UID:         "",
				},
				Spec: &models.V1OpenStackCloudAccount{
					CaCert:           "ca-cert",
					DefaultDomain:    "default-domain",
					DefaultProject:   "default-project",
					IdentityEndpoint: ptr.To("http://identity.endpoint"),
					Insecure:         true,
					ParentRegion:     "parent-region",
					Password:         ptr.To("password"),
					Username:         ptr.To("username"),
				},
			},
		},
		{
			name: "Missing Optional Fields",
			input: map[string]interface{}{
				"name":                     "openstack-account",
				"private_cloud_gateway_id": "private-cloud-gateway-id",
				"default_domain":           "default-domain",
				"default_project":          "default-project",
				"identity_endpoint":        "http://identity.endpoint",
				"parent_region":            "parent-region",
				"openstack_password":       "password",
				"openstack_username":       "username",
			},
			expected: &models.V1OpenStackAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "openstack-account",
					Annotations: map[string]string{OverlordUID: "private-cloud-gateway-id"},
					UID:         "",
				},
				Spec: &models.V1OpenStackCloudAccount{
					DefaultDomain:    "default-domain",
					DefaultProject:   "default-project",
					IdentityEndpoint: ptr.To("http://identity.endpoint"),
					ParentRegion:     "parent-region",
					Password:         ptr.To("password"),
					Username:         ptr.To("username"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a schema.ResourceData instance
			d := schema.TestResourceDataRaw(t, resourceCloudAccountOpenstack().Schema, tt.input)

			// Call the function under test
			result := toOpenStackAccount(d)

			// Perform assertions
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
			assert.Equal(t, tt.expected.Metadata.Annotations[OverlordUID], result.Metadata.Annotations[OverlordUID])
			assert.Equal(t, tt.expected.Spec.CaCert, result.Spec.CaCert)
			assert.Equal(t, tt.expected.Spec.DefaultDomain, result.Spec.DefaultDomain)
			assert.Equal(t, tt.expected.Spec.DefaultProject, result.Spec.DefaultProject)
			assert.Equal(t, tt.expected.Spec.IdentityEndpoint, result.Spec.IdentityEndpoint)
			assert.Equal(t, tt.expected.Spec.Insecure, result.Spec.Insecure)
			assert.Equal(t, tt.expected.Spec.ParentRegion, result.Spec.ParentRegion)
			assert.Equal(t, tt.expected.Spec.Password, result.Spec.Password)
			assert.Equal(t, tt.expected.Spec.Username, result.Spec.Username)
		})
	}
}

func prepareResourceCloudAccountOpenstack() *schema.ResourceData {
	d := resourceCloudAccountOpenstack().TestResourceData()
	d.SetId("test-openstack-account-id-1")
	_ = d.Set("name", "test-openstack-account-1")
	_ = d.Set("context", "project")
	_ = d.Set("private_cloud_gateway_id", "pcg-id")
	_ = d.Set("openstack_username", "test-uname")
	_ = d.Set("openstack_password", "test-pwd")
	_ = d.Set("identity_endpoint", "test-ep")
	_ = d.Set("openstack_allow_insecure", false)
	_ = d.Set("ca_certificate", "test-cert")
	_ = d.Set("parent_region", "test-region1")
	_ = d.Set("default_domain", "test.com")
	_ = d.Set("default_project", "default")

	return d
}

func TestResourceCloudAccountOpenstackCreate(t *testing.T) {
	d := prepareResourceCloudAccountOpenstack()
	ctx := context.Background()
	diags := resourceCloudAccountOpenStackCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-openstack-account-id-1", d.Id())
}

func TestResourceCloudAccountOpenstackRead(t *testing.T) {
	d := prepareResourceCloudAccountOpenstack()
	ctx := context.Background()
	diags := resourceCloudAccountOpenStackRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-openstack-account-id-1", d.Id())
}

func TestResourceCloudAccountOpenstackUpdate(t *testing.T) {
	d := prepareResourceCloudAccountOpenstack()
	ctx := context.Background()
	diags := resourceCloudAccountOpenStackUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-openstack-account-id-1", d.Id())
}

func TestResourceCloudAccountOpenstackDelete(t *testing.T) {
	d := prepareResourceCloudAccountOpenstack()
	ctx := context.Background()
	diags := resourceCloudAccountOpenStackDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-openstack-account-id-1", d.Id())
}
