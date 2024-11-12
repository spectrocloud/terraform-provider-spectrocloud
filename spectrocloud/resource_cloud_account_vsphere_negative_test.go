package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func skipSchemaAttributes(originalSchema map[string]*schema.Schema, keysToRemove []string) map[string]*schema.Schema {
	newSchema := make(map[string]*schema.Schema)
	for key, value := range originalSchema {
		if !contains(keysToRemove, key) {
			newSchema[key] = value
		}
	}
	return newSchema
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestFlattenVsphereCloudAccountAttributes(t *testing.T) {
	// Create a dummy vSphere account
	account := &models.V1VsphereAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "test-account",
			Annotations: map[string]string{
				OverlordUID: "gateway-id",
			},
		},
		Spec: &models.V1VsphereCloudAccount{
			VcenterServer: ptr.To("vcenter.example.com"),
			Username:      ptr.To("user"),
			Insecure:      false,
		},
	}

	// Create a table of test cases
	testCases := []struct {
		AttrName    string
		ExpectedErr bool
	}{
		{"name", false},
		{"context", false},
		{"private_cloud_gateway_id", false},
		{"vsphere_vcenter", false},
		{"vsphere_username", false},
		{"vsphere_ignore_insecure_error", false},
	}

	// Get a copy of the original schema
	originalSchema := resourceCloudAccountVsphere().Schema

	// Iterate through each test case
	for _, test := range testCases {
		attrName := test.AttrName
		expectedErr := test.ExpectedErr

		// Get the attribute from the original schema
		_, ok := originalSchema[attrName]
		if !ok {
			t.Errorf("Attribute %s: Not found in original schema", attrName)
			continue
		}

		d := resourceCloudAccountVsphere().TestResourceData()

		// Test case where d.Set returns an error
		diags, _ := flattenVsphereCloudAccount(d, account)

		if expectedErr {
			if len(diags) != 1 {
				t.Errorf("Attribute %s: Expected one diagnostic, got %d", attrName, len(diags))
			}

			// Check if diags has error for specific attribute
			if !diags.HasError() {
				t.Errorf("attribute %s: Expected error, got no error", attrName)
			}
		} else {
			if len(diags) != 0 {
				t.Errorf("attribute %s: Expected no diagnostics, got %d", attrName, len(diags))
			}
		}
	}
}
