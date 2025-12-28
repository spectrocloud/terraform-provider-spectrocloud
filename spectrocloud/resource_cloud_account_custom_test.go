package spectrocloud

import (
	"context"
	"errors"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestResourceCustomCloudAccount(t *testing.T) {
	// Create a mock resource
	r := resourceCloudAccountCustom()

	// Test CreateContext function
	createCtx := r.CreateContext
	assert.NotNil(t, createCtx)

	// Test ReadContext function
	readCtx := r.ReadContext
	assert.NotNil(t, readCtx)

	// Test UpdateContext function
	updateCtx := r.UpdateContext
	assert.NotNil(t, updateCtx)

	// Test DeleteContext function
	deleteCtx := r.DeleteContext
	assert.NotNil(t, deleteCtx)
}

func TestToCustomCloudAccount(t *testing.T) {
	// Mock resource data
	d := resourceCloudAccountCustom().TestResourceData()
	d.Set("name", "test-name")
	d.Set("cloud", "testcloud")
	d.Set("private_cloud_gateway_id", "test-private-cloud-gateway-id")
	cred := map[string]interface{}{
		"username": "test-username",
		"password": "test-password",
	}
	d.Set("credentials", cred)

	account, err := toCloudAccountCustom(d)

	// Assert that no error occurred during conversion
	assert.NoError(t, err)
	// Assert the metadata
	assert.Equal(t, "test-name", account.Metadata.Name)
	assert.Equal(t, "test-private-cloud-gateway-id", account.Metadata.Annotations[OverlordUID])
	// Assert the credentials
	assert.Equal(t, "test-username", account.Spec.Credentials["username"])
	assert.Equal(t, "test-password", account.Spec.Credentials["password"])
}

func TestFlattenCustomCloudAccount(t *testing.T) {
	// Create a mock resource data
	d := resourceCloudAccountCustom().TestResourceData()
	d.Set("name", "test-name")
	d.Set("cloud", "test-cloud")
	d.Set("private_cloud_gateway_id", "test-private-cloud-gateway-id")
	cred := map[string]interface{}{
		"username": "test-username",
		"password": "test-password",
	}
	d.Set("credentials", cred)
	account := &models.V1CustomAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "test-name",
			Annotations: map[string]string{
				"scope":     "project",
				OverlordUID: "test-private-cloud-gateway-id",
			},
		},
		Kind: "test-cloud",
	}
	diags, hasErrors := flattenCloudAccountCustom(d, account)
	assert.False(t, hasErrors)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-name", d.Get("name"))
	assert.Equal(t, "project", d.Get("context"))
	assert.Equal(t, "test-private-cloud-gateway-id", d.Get("private_cloud_gateway_id"))
	assert.Equal(t, "test-cloud", d.Get("cloud"))
}

// mock
func TestResourceCustomCloudAccountCreate(t *testing.T) {
	// Mock context and resource data
	ctx := context.Background()
	d := resourceCloudAccountCustom().TestResourceData()
	_ = d.Set("name", "test-name")
	_ = d.Set("cloud", "test-cloud")
	_ = d.Set("private_cloud_gateway_id", "test-private-cloud-gateway-id")
	cred := map[string]interface{}{
		"username": "test-username",
		"password": "test-password",
	}
	_ = d.Set("credentials", cred)

	_ = d.Set("context", "test-context")
	_ = d.Set("cloud", "test-cloud")
	diags := resourceCloudAccountCustomCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "mock-uid", d.Id())
}

func TestResourceCustomCloudAccountCreateError(t *testing.T) {
	// Mock context and resource data
	ctx := context.Background()
	d := resourceCloudAccountCustom().TestResourceData()
	_ = d.Set("name", "test-name")
	_ = d.Set("cloud", "test-cloud")
	_ = d.Set("private_cloud_gateway_id", "test-private-cloud-gateway-id")
	cred := map[string]interface{}{
		"username": "test-username",
		"password": "test-password",
	}
	_ = d.Set("credentials", cred)

	// Set up mock client
	_ = d.Set("context", "test-context")
	_ = d.Set("cloud", "test-cloud")
	diags := resourceCloudAccountCustomCreate(ctx, d, unitTestMockAPINegativeClient)
	assert.Error(t, errors.New("unable to find account"))
	assert.Len(t, diags, 1)
	assert.Equal(t, "", d.Id())
}

func TestResourceCustomCloudAccountRead(t *testing.T) {
	ctx := context.Background()
	d := resourceCloudAccountCustom().TestResourceData()

	d.SetId("mock-uid")
	_ = d.Set("context", "test-context")
	_ = d.Set("cloud", "test-cloud")
	diags := resourceCloudAccountCustomRead(ctx, d, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "mock-uid", d.Id())
}

func TestResourceCustomCloudAccountUpdate(t *testing.T) {
	ctx := context.Background()
	d := resourceCloudAccountCustom().TestResourceData()

	d.SetId("existing-id")
	_ = d.Set("name", "test-name")
	_ = d.Set("context", "updated-context")
	_ = d.Set("cloud", "updated-cloud")
	_ = d.Set("private_cloud_gateway_id", "test-private-cloud-gateway-id")
	cred := map[string]interface{}{
		"username": "test-username",
		"password": "test-password",
	}
	_ = d.Set("credentials", cred)
	diags := resourceCloudAccountCustomUpdate(ctx, d, unitTestMockAPIClient)

	assert.Len(t, diags, 0)
}

func TestResourceCustomCloudAccountDelete(t *testing.T) {
	ctx := context.Background()
	d := resourceCloudAccountCustom().TestResourceData()

	d.SetId("existing-id")
	_ = d.Set("context", "test-context")
	_ = d.Set("cloud", "test-cloud")
	diags := resourceCloudAccountCustomDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}
