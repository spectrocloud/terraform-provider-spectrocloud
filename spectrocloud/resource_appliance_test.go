package spectrocloud

import (
	"context"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"

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
