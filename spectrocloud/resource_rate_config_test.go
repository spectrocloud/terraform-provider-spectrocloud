package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func publicCloudRateBlock(computeOptimizedCompute, computeOptimizedMemory, memoryOptimizedCompute, memoryOptimizedMemory float64) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"compute_optimized": []interface{}{
				map[string]interface{}{
					"compute_rate_proportion": computeOptimizedCompute,
					"memory_rate_proportion":  computeOptimizedMemory,
				},
			},
			"memory_optimized": []interface{}{
				map[string]interface{}{
					"compute_rate_proportion": memoryOptimizedCompute,
					"memory_rate_proportion":  memoryOptimizedMemory,
				},
			},
		},
	}
}

func privateCloudRateBlock(cpu, gpu, memory, storage float64) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"cpu_unit_price_per_hour":         cpu,
			"gpu_unit_price_per_hour":         gpu,
			"memory_unit_price_gib_per_hour":  memory,
			"storage_unit_price_gib_per_hour": storage,
		},
	}
}

func prepareRateConfigResourceData() *schema.ResourceData {
	d := resourceRateConfig().TestResourceData()
	_ = d.Set("aws", publicCloudRateBlock(60, 40, 30, 70))
	_ = d.Set("vsphere", privateCloudRateBlock(0.05, 3.5, 0.01, 0.0001))
	_ = d.Set("custom", []interface{}{
		map[string]interface{}{
			"cloud_type":  "test-custom-cloud",
			"rate_config": privateCloudRateBlock(0.07, 4.5, 0.02, 0.0002),
		},
	})
	return d
}

// ---------------------------------------------------------------------------
// Expand
// ---------------------------------------------------------------------------

func TestToRateConfig(t *testing.T) {
	d := prepareRateConfigResourceData()

	rateConfig := toRateConfig(d)

	require.NotNil(t, rateConfig)
	require.NotNil(t, rateConfig.Aws)
	assert.Equal(t, float32(60), rateConfig.Aws.ComputeOptimized.ComputeRateProportion)
	assert.Equal(t, float32(40), rateConfig.Aws.ComputeOptimized.MemoryRateProportion)
	assert.Equal(t, float32(30), rateConfig.Aws.MemoryOptimized.ComputeRateProportion)
	assert.Equal(t, float32(70), rateConfig.Aws.MemoryOptimized.MemoryRateProportion)

	require.NotNil(t, rateConfig.Vsphere)
	assert.Equal(t, 0.05, rateConfig.Vsphere.CPUUnitPricePerHour)
	assert.Equal(t, 3.5, rateConfig.Vsphere.GpuUnitPricePerHour)
	assert.Equal(t, 0.01, rateConfig.Vsphere.MemoryUnitPriceGiBPerHour)
	assert.Equal(t, 0.0001, rateConfig.Vsphere.StorageUnitPriceGiBPerHour)

	// Clouds absent from config are left nil; Palette defaults them server-side.
	assert.Nil(t, rateConfig.Azure)
	assert.Nil(t, rateConfig.Gcp)
	assert.Nil(t, rateConfig.Maas)

	require.Len(t, rateConfig.Custom, 1)
	assert.Equal(t, "test-custom-cloud", rateConfig.Custom[0].CloudType)
	assert.Equal(t, 0.07, rateConfig.Custom[0].RateConfig.CPUUnitPricePerHour)
}

func TestToRateConfigEmpty(t *testing.T) {
	rateConfig := toRateConfig(resourceRateConfig().TestResourceData())

	require.NotNil(t, rateConfig)
	assert.Nil(t, rateConfig.Aws)
	assert.Nil(t, rateConfig.Vsphere)
	// Custom has no omitempty, so it must never marshal as null.
	assert.NotNil(t, rateConfig.Custom)
	assert.Empty(t, rateConfig.Custom)
}

func TestToRateConfigDefault(t *testing.T) {
	rateConfig := toRateConfigDefault()

	require.NotNil(t, rateConfig)
	for name, cloud := range map[string]*models.V1PublicCloudRateConfig{
		"aws": rateConfig.Aws, "azure": rateConfig.Azure, "gcp": rateConfig.Gcp,
	} {
		require.NotNil(t, cloud, name)
		assert.Equal(t, float32(65), cloud.ComputeOptimized.ComputeRateProportion, name)
		assert.Equal(t, float32(35), cloud.ComputeOptimized.MemoryRateProportion, name)
		assert.Equal(t, float32(25), cloud.MemoryOptimized.ComputeRateProportion, name)
		assert.Equal(t, float32(75), cloud.MemoryOptimized.MemoryRateProportion, name)
	}
	for name, cloud := range map[string]*models.V1PrivateCloudRateConfig{
		"vsphere": rateConfig.Vsphere, "maas": rateConfig.Maas, "edge": rateConfig.Edge,
		"edge_native": rateConfig.EdgeNative, "generic": rateConfig.Generic,
		"apache_cloudstack": rateConfig.ApacheCloudstack,
	} {
		require.NotNil(t, cloud, name)
		assert.Equal(t, 0.021811, cloud.CPUUnitPricePerHour, name)
		assert.Equal(t, 2.933908, cloud.GpuUnitPricePerHour, name)
		assert.Equal(t, 0.002923, cloud.MemoryUnitPriceGiBPerHour, name)
		assert.Equal(t, 0.000028, cloud.StorageUnitPriceGiBPerHour, name)
	}
	assert.NotNil(t, rateConfig.Custom)
	assert.Empty(t, rateConfig.Custom)
}

// Palette dereferences both instance categories without a nil check, so the
// expander must never emit a nil ComputeOptimized/MemoryOptimized.
func TestToPublicCloudRateConfigAlwaysEmitsBothCategories(t *testing.T) {
	publicCloud := toPublicCloudRateConfig([]interface{}{map[string]interface{}{}})

	require.NotNil(t, publicCloud)
	require.NotNil(t, publicCloud.ComputeOptimized)
	require.NotNil(t, publicCloud.MemoryOptimized)
}

func TestToRateConfigExpandersNilSafe(t *testing.T) {
	assert.Nil(t, toPublicCloudRateConfig(nil))
	assert.Nil(t, toPublicCloudRateConfig([]interface{}{}))
	assert.Nil(t, toPublicCloudRateConfig([]interface{}{nil}))
	assert.Nil(t, toPrivateCloudRateConfig(nil))
	assert.Nil(t, toPrivateCloudRateConfig([]interface{}{}))
	assert.Empty(t, toCustomCloudRateConfigs(nil))
	assert.Empty(t, toCustomCloudRateConfigs("not-a-list"))
	// A custom entry without rates is dropped rather than sent as null.
	assert.Empty(t, toCustomCloudRateConfigs([]interface{}{
		map[string]interface{}{"cloud_type": "test-custom-cloud"},
	}))
}

// ---------------------------------------------------------------------------
// Flatten
// ---------------------------------------------------------------------------

func TestFlattenRateConfig(t *testing.T) {
	d := resourceRateConfig().TestResourceData()
	rateConfig := &models.V1RateConfig{
		Aws: &models.V1PublicCloudRateConfig{
			ComputeOptimized: &models.V1CloudInstanceRateConfig{ComputeRateProportion: 65, MemoryRateProportion: 35},
			MemoryOptimized:  &models.V1CloudInstanceRateConfig{ComputeRateProportion: 25, MemoryRateProportion: 75},
		},
		Vsphere: &models.V1PrivateCloudRateConfig{
			CPUUnitPricePerHour:        0.021811,
			GpuUnitPricePerHour:        2.933908,
			MemoryUnitPriceGiBPerHour:  0.002923,
			StorageUnitPriceGiBPerHour: 0.000028,
		},
		Custom: []*models.V1CustomCloudRateConfig{
			{
				CloudType:  "test-custom-cloud",
				RateConfig: &models.V1PrivateCloudRateConfig{CPUUnitPricePerHour: 0.07},
			},
		},
	}

	require.NoError(t, flattenRateConfig(rateConfig, d))

	aws := d.Get("aws").([]interface{})
	require.Len(t, aws, 1)
	computeOptimized := aws[0].(map[string]interface{})["compute_optimized"].([]interface{})
	require.Len(t, computeOptimized, 1)
	assert.Equal(t, 65.0, computeOptimized[0].(map[string]interface{})["compute_rate_proportion"])
	assert.Equal(t, 35.0, computeOptimized[0].(map[string]interface{})["memory_rate_proportion"])

	vsphere := d.Get("vsphere").([]interface{})
	require.Len(t, vsphere, 1)
	assert.Equal(t, 0.021811, vsphere[0].(map[string]interface{})["cpu_unit_price_per_hour"])
	assert.Equal(t, 2.933908, vsphere[0].(map[string]interface{})["gpu_unit_price_per_hour"])

	custom := d.Get("custom").([]interface{})
	require.Len(t, custom, 1)
	assert.Equal(t, "test-custom-cloud", custom[0].(map[string]interface{})["cloud_type"])

	// Clouds missing from the payload flatten to empty blocks, not errors.
	assert.Empty(t, d.Get("azure").([]interface{}))
	assert.Empty(t, d.Get("maas").([]interface{}))
}

func TestFlattenRateConfigNilSafe(t *testing.T) {
	d := resourceRateConfig().TestResourceData()
	assert.NoError(t, flattenRateConfig(nil, d))
	assert.Empty(t, flattenPublicCloudRateConfig(nil))
	assert.Empty(t, flattenPrivateCloudRateConfig(nil))
	assert.Empty(t, flattenCloudInstanceRateConfig(nil))
	assert.Empty(t, flattenCustomCloudRateConfigs(nil))
	assert.Empty(t, flattenCustomCloudRateConfigs([]*models.V1CustomCloudRateConfig{nil}))
}

func TestRateConfigExpandFlattenRoundTrip(t *testing.T) {
	d := prepareRateConfigResourceData()
	expanded := toRateConfig(d)

	roundTripped := resourceRateConfig().TestResourceData()
	require.NoError(t, flattenRateConfig(expanded, roundTripped))

	assert.Equal(t, d.Get("aws"), roundTripped.Get("aws"))
	assert.Equal(t, d.Get("vsphere"), roundTripped.Get("vsphere"))
	assert.Equal(t, d.Get("custom"), roundTripped.Get("custom"))
}

// ---------------------------------------------------------------------------
// CustomizeDiff
// ---------------------------------------------------------------------------

func TestValidateRateProportions(t *testing.T) {
	t.Run("accepts proportions adding up to 100", func(t *testing.T) {
		d := resourceRateConfig().TestResourceData()
		_ = d.Set("gcp", publicCloudRateBlock(65, 35, 25, 75))
		assert.NoError(t, validateRateProportionsIn(d))
	})

	t.Run("rejects proportions not adding up to 100", func(t *testing.T) {
		d := resourceRateConfig().TestResourceData()
		_ = d.Set("azure", publicCloudRateBlock(65, 35, 10, 20))
		err := validateRateProportionsIn(d)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "azure.memory_optimized")
		assert.Contains(t, err.Error(), "must add up to 100")
	})

	t.Run("ignores clouds that are not configured", func(t *testing.T) {
		d := resourceRateConfig().TestResourceData()
		_ = d.Set("vsphere", privateCloudRateBlock(0.05, 3.5, 0.01, 0.0001))
		assert.NoError(t, validateRateProportionsIn(d))
	})
}

// TestResourceRateConfigCustomizeDiff drives the validateRateProportions
// wrapper through a real Resource.Diff cycle, so the wiring (not just the
// already-tested validateRateProportionsIn helper) is covered.
func TestResourceRateConfigCustomizeDiff(t *testing.T) {
	r := resourceRateConfig()
	require.NotNil(t, r.CustomizeDiff, "rate config CustomizeDiff must be wired")

	rateConfigDiff := func(cfg map[string]interface{}) error {
		_, err := r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(cfg), unitTestMockAPIClient)
		return err
	}

	assert.NoError(t, rateConfigDiff(map[string]interface{}{
		"aws": publicCloudRateBlock(65, 35, 25, 75),
	}))

	err := rateConfigDiff(map[string]interface{}{
		"aws": publicCloudRateBlock(65, 35, 10, 20),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must add up to 100")
}

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func TestResourceRateConfigCRUD(t *testing.T) {
	testResourceCRUD(t, prepareRateConfigResourceData, unitTestMockAPIClient,
		resourceRateConfigCreate, resourceRateConfigRead,
		resourceRateConfigUpdate, resourceRateConfigDelete)
}

func TestResourceRateConfigCreateSetsSingletonID(t *testing.T) {
	d := prepareRateConfigResourceData()
	diags := resourceRateConfigCreate(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, rateConfigID, d.Id())
}

func TestResourceRateConfigDeleteClearsID(t *testing.T) {
	d := prepareRateConfigResourceData()
	d.SetId(rateConfigID)
	diags := resourceRateConfigDelete(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Empty(t, d.Id())
}

func TestResourceRateConfigReadWithoutID(t *testing.T) {
	// Singleton guard: a non-canonical ID clears the resource rather than
	// flattening state that belongs to a different object.
	d := prepareRateConfigResourceData()
	d.SetId("some-random-id")
	diags := resourceRateConfigRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Empty(t, d.Id(), "Read must clear a non-canonical ID")
}

func TestResourceRateConfigReadPopulatesState(t *testing.T) {
	d := resourceRateConfig().TestResourceData()
	d.SetId(rateConfigID)
	diags := resourceRateConfigRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Len(t, d.Get("aws").([]interface{}), 1)
	assert.Len(t, d.Get("vsphere").([]interface{}), 1)
	assert.Len(t, d.Get("custom").([]interface{}), 1)
}

func TestResourceRateConfigCRUDNegative(t *testing.T) {
	for _, op := range []string{"Create", "Read", "Update", "Delete"} {
		t.Run(op+" surfaces API error", func(t *testing.T) {
			testResourceCRUDNegative(t, op, prepareRateConfigResourceData,
				unitTestMockAPINegativeClient,
				resourceRateConfigCreate, resourceRateConfigRead,
				resourceRateConfigUpdate, resourceRateConfigDelete,
				op != "Create", "Invalid rate config")
		})
	}
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

func TestResourceRateConfigImport(t *testing.T) {
	t.Run("uid matches tenant", func(t *testing.T) {
		d := resourceRateConfig().TestResourceData()
		d.SetId("test-tenant-uid") // matches getMockUserInfoPayload().tenantUid
		got, err := resourceRateConfigImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, rateConfigID, got[0].Id())
		assert.Len(t, got[0].Get("vsphere").([]interface{}), 1)
	})

	t.Run("uid mismatch errors", func(t *testing.T) {
		d := resourceRateConfig().TestResourceData()
		d.SetId("some-other-tenant-uid")
		_, err := resourceRateConfigImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})
}
