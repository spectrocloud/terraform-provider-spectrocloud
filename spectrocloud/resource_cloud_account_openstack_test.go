package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
	"testing"
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
					IdentityEndpoint: types.Ptr("http://identity.endpoint"),
					Insecure:         true,
					ParentRegion:     "parent-region",
					Password:         types.Ptr("password"),
					Username:         types.Ptr("username"),
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
					IdentityEndpoint: types.Ptr("http://identity.endpoint"),
					ParentRegion:     "parent-region",
					Password:         types.Ptr("password"),
					Username:         types.Ptr("username"),
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
