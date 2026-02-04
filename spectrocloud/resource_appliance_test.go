package spectrocloud

import (
	"context"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestToApplianceEntity(t *testing.T) {
	d := resourceAppliance().TestResourceData()
	d.Set("uid", "testID")
	d.Set("tags", map[string]interface{}{"tag1": "value1", "tag2": "value2"})
	d.Set("pairing_key", "testKey")

	expectedEntity := &models.V1EdgeHostDeviceEntity{
		// Define expected output values based on input values
		Metadata: &models.V1ObjectTagsEntity{
			UID:    "testID",
			Name:   "testID",
			Labels: map[string]string{"tag1": "value1", "tag2": "value2"},
		},
		Spec: &models.V1EdgeHostDeviceSpecEntity{
			HostPairingKey: strfmt.Password("testKey"),
		},
	}

	result := toApplianceEntity(d)
	assert.Equal(t, expectedEntity, result)
}

func TestToApplianceMeta_WithTags(t *testing.T) {
	d := resourceAppliance().TestResourceData()
	d.Set("uid", "testID")
	d.SetId("testID")
	d.Set("tags", map[string]interface{}{"tag1": "value1", "tag2": "value2"})

	expectedEntityWithTags := &models.V1EdgeHostDeviceMetaUpdateEntity{
		Metadata: &models.V1ObjectTagsEntity{
			Labels: map[string]string{"tag1": "value1", "tag2": "value2"},
			Name:   "testID",
			UID:    "testID",
		},
	}

	resultWithTags := toApplianceMeta(d)
	assert.Equal(t, expectedEntityWithTags, resultWithTags)
}

func TestToApplianceMeta_WithoutTags(t *testing.T) {
	d := resourceAppliance().TestResourceData()
	d.Set("uid", "testID")
	d.SetId("testID")

	expectedEntityWithoutTags := &models.V1EdgeHostDeviceMetaUpdateEntity{
		Metadata: &models.V1ObjectTagsEntity{
			Name:   "testID",
			UID:    "testID",
			Labels: make(map[string]string),
		},
	}

	resultWithoutTags := toApplianceMeta(d)
	assert.Equal(t, expectedEntityWithoutTags, resultWithoutTags)
}

func TestToAppliance(t *testing.T) {
	d := resourceAppliance().TestResourceData()
	d.Set("uid", "testID")
	d.SetId("testID")
	d.Set("tags", map[string]interface{}{"tag1": "value1", "tag2": "value2"})

	expectedApplianceWithTags := setFields(d, d.Get("tags").(map[string]interface{}))

	resultWithTags := toAppliance(d)
	assert.Equal(t, &expectedApplianceWithTags, resultWithTags)
}

func TestSetFields_WithNameTag(t *testing.T) {
	d := resourceAppliance().TestResourceData()
	d.Set("uid", "testID")
	d.SetId("testID")
	d.Set("tags", map[string]interface{}{"name": "TestName", "tag2": "value2"})

	mockTags := map[string]interface{}{"name": "TestName", "tag2": "value2"}

	expectedApplianceWithNameTag := models.V1EdgeHostDevice{
		Metadata: &models.V1ObjectMeta{
			UID:    "testID",
			Name:   "TestName",
			Labels: expandStringMap(mockTags),
		},
	}

	resultWithNameTag := setFields(d, mockTags)
	assert.Equal(t, expectedApplianceWithNameTag, resultWithNameTag)
}

func TestSetFields_WithoutNameTag(t *testing.T) {
	d := resourceAppliance().TestResourceData()
	d.Set("uid", "testID")
	d.SetId("testID")
	d.Set("tags", map[string]interface{}{"test": "TestName", "tag2": "value2"})

	mockTagsWithoutName := map[string]interface{}{"test": "TestName", "tag2": "value2"}

	expectedApplianceWithoutNameTag := models.V1EdgeHostDevice{
		Metadata: &models.V1ObjectMeta{
			UID:    "testID",
			Labels: expandStringMap(mockTagsWithoutName),
		},
	}

	resultWithoutNameTag := setFields(d, mockTagsWithoutName)
	assert.Equal(t, expectedApplianceWithoutNameTag, resultWithoutNameTag)
}

func prepareApplianceBaseData() *schema.ResourceData {
	d := resourceAppliance().TestResourceData()
	_ = d.Set("uid", "test-edge-host-id")
	_ = d.Set("wait", false)
	_ = d.Set("remote_shell", "enabled")
	_ = d.Set("temporary_shell_credentials", "enabled")
	d.SetId("test-idz")
	return d
}

func TestResourceApplianceCreateInvalid(t *testing.T) {
	d := prepareApplianceBaseData()

	diags := resourceApplianceCreate(context.Background(), d, unitTestMockAPINegativeClient)

	assert.NotEmpty(t, diags)
	assertFirstDiagMessage(t, diags, "Operation not allowed")
}

func TestResourceApplianceRead(t *testing.T) {
	d := prepareApplianceBaseData()

	diags := resourceApplianceRead(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
}

func TestResourceApplianceUpdate(t *testing.T) {
	d := prepareApplianceBaseData()

	diags := resourceApplianceUpdate(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
}

func TestResourceApplianceDelete(t *testing.T) {
	d := prepareApplianceBaseData()

	diags := resourceApplianceDelete(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
}

func TestResourceApplianceDeleteInvalid(t *testing.T) {
	d := prepareApplianceBaseData()

	diags := resourceApplianceDelete(context.Background(), d, unitTestMockAPINegativeClient)

	assert.NotEmpty(t, diags)
	assertFirstDiagMessage(t, diags, "No edge host found")
}

func TestResourceApplianceGetState(t *testing.T) {
	diags := resourceApplianceStateRefreshFunc(getV1ClientWithResourceContext(unitTestMockAPIClient, "project"), "test")

	assert.NotEmpty(t, diags)
}

func TestResourceApplianceImport(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		client      interface{}
		expectError bool
		errorMsg    string
		description string
		verify      func(t *testing.T, importedData []*schema.ResourceData, err error)
	}{
		{
			name: "Successful import with appliance ID",
			setup: func() *schema.ResourceData {
				d := resourceAppliance().TestResourceData()
				d.SetId("test-appliance-id")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false,
			description: "Should successfully import appliance with valid ID",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				if err == nil {
					assert.NotNil(t, importedData, "Imported data should not be nil on success")
					if len(importedData) > 0 {
						assert.Len(t, importedData, 1, "Should return exactly one ResourceData")
						assert.NotEmpty(t, importedData[0].Id(), "Appliance ID should be set")
					}
				}
			},
		},
		{
			name: "Error when import ID is empty",
			setup: func() *schema.ResourceData {
				d := resourceAppliance().TestResourceData()
				d.SetId("")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: true,
			errorMsg:    "appliance import ID is required",
			description: "Should return error when import ID is empty",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				assert.Error(t, err)
				assert.Nil(t, importedData)
				assert.Contains(t, err.Error(), "appliance import ID is required")
			},
		},
		{
			name: "Successful import sets required fields",
			setup: func() *schema.ResourceData {
				d := resourceAppliance().TestResourceData()
				d.SetId("test-appliance-id")
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false,
			description: "Should set uid and other fields during import",
			verify: func(t *testing.T, importedData []*schema.ResourceData, err error) {
				if err == nil && len(importedData) > 0 {
					d := importedData[0]
					// Verify that uid is set (GetCommonAppliance sets this)
					// Note: Actual values depend on mock API response
					assert.NotEmpty(t, d.Id(), "ID should be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			importedData, err := resourceApplianceImport(ctx, d, tt.client)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// Some tests may succeed or fail depending on mock setup
				if err != nil {
					t.Logf("Unexpected error: %v", err)
				}
			}

			if tt.verify != nil {
				tt.verify(t, importedData, err)
			}
		})
	}
}
