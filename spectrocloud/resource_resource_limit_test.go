package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToResourceLimits(t *testing.T) {
	d := resourceResourceLimit().TestResourceData()

	// Set custom values for testing
	d.Set("alert", 50)
	d.Set("api_keys", 10)
	d.Set("application_deployment", 80)

	resourceLimits, err := toResourceLimits(d)
	assert.NoError(t, err)
	assert.NotNil(t, resourceLimits)

	// Verify specific values
	assert.Equal(t, int64(50), resourceLimits.Resources[0].Limit)
	assert.Equal(t, int64(10), resourceLimits.Resources[1].Limit)
	assert.Equal(t, int64(80), resourceLimits.Resources[2].Limit)
}

func TestToResourceDefaultLimits(t *testing.T) {
	d := resourceResourceLimit().TestResourceData()

	resourceLimits, err := toResourceDefaultLimits(d)
	assert.NoError(t, err)
	assert.NotNil(t, resourceLimits)

	// Verify default values from KindToFieldMapping
	for i, mapping := range KindToFieldMapping {
		assert.Equal(t, mapping.Default, resourceLimits.Resources[i].Limit)
	}
}

func TestFlattenResourceLimits(t *testing.T) {
	d := resourceResourceLimit().TestResourceData()

	resourceLimits := &models.V1TenantResourceLimits{
		Resources: []*models.V1TenantResourceLimit{
			{Kind: models.V1ResourceLimitTypeAlert, Limit: 75},
			{Kind: models.V1ResourceLimitTypeAPIKey, Limit: 15},
			{Kind: models.V1ResourceLimitTypeAppdeployment, Limit: 90},
		},
	}

	err := flattenResourceLimits(resourceLimits, d)
	assert.NoError(t, err)

	// Verify the values were correctly set
	assert.Equal(t, 75, d.Get("alert"))
	assert.Equal(t, 15, d.Get("api_keys"))
	assert.Equal(t, 90, d.Get("application_deployment"))
}
