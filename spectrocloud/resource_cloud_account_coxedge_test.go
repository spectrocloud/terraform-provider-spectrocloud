package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/hapi/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// Test for toCoxEdgeAccount
func TestToCoxEdgeAccount(t *testing.T) {
	rd := resourceCloudAccountCoxEdge().TestResourceData() // Assuming this method exists
	rd.Set("name", "coxedge_unit_test_acc")
	rd.Set("api_base_url", "https://coxedge.api.example.com")
	rd.Set("api_key", "test_api_key")
	rd.Set("environment", "test_environment")
	rd.Set("organization_id", "test_org_id")
	rd.Set("service", "test_service")

	acc := toCoxEdgeAccount(rd)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("api_base_url"), *acc.Spec.APIBaseURL)
	assert.Equal(t, rd.Get("api_key"), *acc.Spec.APIKey)
	assert.Equal(t, rd.Get("environment"), acc.Spec.Environment)
	assert.Equal(t, rd.Get("organization_id"), acc.Spec.OrganizationID)
	assert.Equal(t, rd.Get("service"), acc.Spec.Service)
}

// Test for flattenCoxEdgeCloudAccount
func TestFlattenCoxEdgeCloudAccount(t *testing.T) {
	rd := resourceCloudAccountCoxEdge().TestResourceData() // Assuming this method exists
	account := &models.V1CoxEdgeAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "test_account",
			UID:  "abcdef",
		},
		Spec: &models.V1CoxEdgeCloudAccount{
			APIBaseURL:     types.Ptr("https://coxedge.api.example.com"),
			APIKey:         types.Ptr("test_api_key"),
			Environment:    "test_environment",
			OrganizationID: "test_org_id",
			Service:        "test_service",
		},
	}

	diags, hasError := flattenCoxEdgeCloudAccount(rd, account)

	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "test_account", rd.Get("name"))
	assert.Equal(t, "https://coxedge.api.example.com", rd.Get("api_base_url"))
	assert.Equal(t, "test_environment", rd.Get("environment"))
	assert.Equal(t, "test_org_id", rd.Get("organization_id"))
	assert.Equal(t, "test_service", rd.Get("service"))
}
