package spectrocloud

import (
	"context"
	"errors"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResourceCustomCloudAccount(t *testing.T) {
	// Create a mock resource
	r := resourceCustomCloudAccount()

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
	d := resourceCustomCloudAccount().TestResourceData()
	d.Set("name", "test-name")
	d.Set("cloud", "testcloud")
	d.Set("private_cloud_gateway_id", "test-private-cloud-gateway-id")
	cred := map[string]interface{}{
		"username": "test-username",
		"password": "test-password",
	}
	d.Set("credentials", cred)

	account, err := toCustomCloudAccount(d)

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
	d := resourceCustomCloudAccount().TestResourceData()
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
	diags, hasErrors := flattenCustomCloudAccount(d, account)
	assert.False(t, hasErrors)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-name", d.Get("name"))
	assert.Equal(t, "project", d.Get("context"))
	assert.Equal(t, "test-private-cloud-gateway-id", d.Get("private_cloud_gateway_id"))
	assert.Equal(t, "test-cloud", d.Get("cloud"))
}

func TestResourceCustomCloudAccountCreate(t *testing.T) {
	// Mock context and resource data
	ctx := context.Background()
	d := resourceCustomCloudAccount().TestResourceData()
	d.Set("name", "test-name")
	d.Set("cloud", "test-cloud")
	d.Set("private_cloud_gateway_id", "test-private-cloud-gateway-id")
	cred := map[string]interface{}{
		"username": "test-username",
		"password": "test-password",
	}
	d.Set("credentials", cred)

	mockClient := &client.V1Client{
		ValidateCustomCloudTypeFn: func(cloudType, accountContext string) error {
			return nil
		},
		CreateCustomCloudAccountFn: func(account *models.V1CustomAccountEntity, cloudType, accountContext string) (string, error) {
			return "mock-uid", nil
		},
		GetCustomCloudAccountFn: func(uid, cloudType string, accountContext string) (*models.V1CustomAccount, error) {
			return &models.V1CustomAccount{
				Kind: "test-cloud",
				Metadata: &models.V1ObjectMeta{
					Annotations: map[string]string{
						OverlordUID: "test-private-cloud-gateway-id",
					},
					UID: "mock-uid",
				},
				Spec: &models.V1CustomCloudAccount{
					Credentials: map[string]string{
						"username": "test-username",
						"password": "test-password",
					},
				},
			}, nil
		},
	}
	d.Set("context", "test-context")
	d.Set("cloud", "test-cloud")
	diags := resourceCustomCloudAccountCreate(ctx, d, mockClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "mock-uid", d.Id())
}

func TestResourceCustomCloudAccountCreateError(t *testing.T) {
	// Mock context and resource data
	ctx := context.Background()
	d := resourceCustomCloudAccount().TestResourceData()
	d.Set("name", "test-name")
	d.Set("cloud", "test-cloud")
	d.Set("private_cloud_gateway_id", "test-private-cloud-gateway-id")
	cred := map[string]interface{}{
		"username": "test-username",
		"password": "test-password",
	}
	d.Set("credentials", cred)

	// Set up mock client
	mockClient := &client.V1Client{
		ValidateCustomCloudTypeFn: func(cloudType, accountContext string) error {
			return nil
		},
		CreateCustomCloudAccountFn: func(account *models.V1CustomAccountEntity, cloudType, accountContext string) (string, error) {
			return "", errors.New("unable to find account")
		},
		GetCustomCloudAccountFn: func(uid, cloudType string, accountContext string) (*models.V1CustomAccount, error) {
			return nil, nil
		},
	}
	d.Set("context", "test-context")
	d.Set("cloud", "test-cloud")
	diags := resourceCustomCloudAccountCreate(ctx, d, mockClient)
	assert.Error(t, errors.New("unable to find account"))
	assert.Len(t, diags, 1)
	assert.Equal(t, "", d.Id())
}

func TestResourceCustomCloudAccountRead(t *testing.T) {
	ctx := context.Background()
	d := resourceCustomCloudAccount().TestResourceData()

	mockClient := &client.V1Client{
		GetCustomCloudAccountFn: func(id, cloudType, accountContext string) (*models.V1CustomAccount, error) {
			if id == "existing-id" {
				return &models.V1CustomAccount{
					Metadata: &models.V1ObjectMeta{
						Name: "test-name",
						Annotations: map[string]string{
							"scope":     "test-scope",
							OverlordUID: "test-overlord-uid",
						},
					},
					Kind: "test-cloud",
				}, nil
			}
			return nil, nil
		},
	}

	d.SetId("existing-id")
	d.Set("context", "test-context")
	d.Set("cloud", "test-cloud")
	diags := resourceCustomCloudAccountRead(ctx, d, mockClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "existing-id", d.Id())
	assert.Equal(t, "test-name", d.Get("name"))
	assert.Equal(t, "test-scope", d.Get("context"))
	assert.Equal(t, "test-overlord-uid", d.Get("private_cloud_gateway_id"))
	assert.Equal(t, "test-cloud", d.Get("cloud"))
}

func TestResourceCustomCloudAccountUpdate(t *testing.T) {
	ctx := context.Background()
	d := resourceCustomCloudAccount().TestResourceData()
	mockClient := &client.V1Client{
		UpdateCustomCloudAccountFn: func(id string, account *models.V1CustomAccountEntity, cloudType, accountContext string) error {
			return nil
		},
		GetCustomCloudAccountFn: func(id, cloudType, accountContext string) (*models.V1CustomAccount, error) {
			return &models.V1CustomAccount{
				Metadata: &models.V1ObjectMeta{
					Name: "updated-name",
					Annotations: map[string]string{
						"scope":     "updated-scope",
						OverlordUID: "updated-overlord-uid",
					},
				},
				Kind: "updated-cloud",
			}, nil
		},
	}

	d.SetId("existing-id")
	d.Set("context", "updated-context")
	d.Set("cloud", "updated-cloud")
	diags := resourceCustomCloudAccountUpdate(ctx, d, mockClient)

	assert.Len(t, diags, 0)
	assert.Equal(t, "existing-id", d.Id())
	assert.Equal(t, "updated-name", d.Get("name"))
	assert.Equal(t, "updated-scope", d.Get("context"))
	assert.Equal(t, "updated-overlord-uid", d.Get("private_cloud_gateway_id"))
	assert.Equal(t, "updated-cloud", d.Get("cloud"))
}

func TestResourceCustomCloudAccountDelete(t *testing.T) {
	ctx := context.Background()
	d := resourceCustomCloudAccount().TestResourceData()
	mockClient := &client.V1Client{
		DeleteCustomCloudAccountFn: func(id, cloudType, accountContext string) error {
			return nil
		},
	}
	d.SetId("existing-id")
	d.Set("context", "test-context")
	d.Set("cloud", "test-cloud")
	diags := resourceCustomCloudAccountDelete(ctx, d, mockClient)
	assert.Len(t, diags, 0)
}
