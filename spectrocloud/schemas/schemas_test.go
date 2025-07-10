package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAppPackSchema(t *testing.T) {
	s := AppPackSchema()

	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, true, s.Required)
	assert.Equal(t, "A list of packs to be applied to the application profile.", s.Description)

	elemSchema, ok := s.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, elemSchema)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["type"].Type)
	assert.Equal(t, true, elemSchema.Schema["type"].Optional)
	assert.Equal(t, "The type of Pack. Allowed values are `container`, `helm`, `manifest`, or `operator-instance`.", elemSchema.Schema["type"].Description)
	assert.Equal(t, "spectro", elemSchema.Schema["type"].Default)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["source_app_tier"].Type)
	assert.Equal(t, true, elemSchema.Schema["source_app_tier"].Optional)
	assert.Equal(t, "The unique id of the pack to be used as the source for the pack.", elemSchema.Schema["source_app_tier"].Description)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["registry_uid"].Type)
	assert.Equal(t, true, elemSchema.Schema["registry_uid"].Optional)
	assert.Equal(t, true, elemSchema.Schema["registry_uid"].Computed)
	assert.Equal(t, "The unique id of the registry to be used for the pack.", elemSchema.Schema["registry_uid"].Description)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["uid"].Type)
	assert.Equal(t, true, elemSchema.Schema["uid"].Optional)
	assert.Equal(t, true, elemSchema.Schema["uid"].Computed)
	assert.Equal(t, "The unique id of the pack. This is a computed field and is not required to be set.", elemSchema.Schema["uid"].Description)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["name"].Type)
	assert.Equal(t, true, elemSchema.Schema["name"].Required)
	assert.Equal(t, "The name of the specified pack.", elemSchema.Schema["name"].Description)

	assert.Equal(t, schema.TypeMap, elemSchema.Schema["properties"].Type)
	assert.Equal(t, true, elemSchema.Schema["properties"].Optional)
	assert.Equal(t, "The various properties required by different database tiers eg: `databaseName` and `databaseVolumeSize` size for Redis etc.", elemSchema.Schema["properties"].Description)

	assert.Equal(t, schema.TypeInt, elemSchema.Schema["install_order"].Type)
	assert.Equal(t, true, elemSchema.Schema["install_order"].Optional)
	assert.Equal(t, 0, elemSchema.Schema["install_order"].Default)
	assert.Equal(t, "The installation priority order of the app profile. The order of priority goes from lowest number to highest number. For example, a value of `-3` would be installed before an app profile with a higher number value. No upper and lower limits exist, and you may specify positive and negative integers. The default value is `0`. ", elemSchema.Schema["install_order"].Description)

	assert.Equal(t, schema.TypeList, elemSchema.Schema["manifest"].Type)
	assert.Equal(t, true, elemSchema.Schema["manifest"].Optional)
	assert.Equal(t, "The manifest of the pack.", elemSchema.Schema["manifest"].Description)

	manifestElemSchema, ok := elemSchema.Schema["manifest"].Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, manifestElemSchema)

	assert.Equal(t, schema.TypeString, manifestElemSchema.Schema["uid"].Type)
	assert.Equal(t, true, manifestElemSchema.Schema["uid"].Computed)

	assert.Equal(t, schema.TypeString, manifestElemSchema.Schema["name"].Type)
	assert.Equal(t, true, manifestElemSchema.Schema["name"].Required)
	assert.Equal(t, "The name of the manifest.", manifestElemSchema.Schema["name"].Description)

	assert.Equal(t, schema.TypeString, manifestElemSchema.Schema["content"].Type)
	assert.Equal(t, true, manifestElemSchema.Schema["content"].Required)
	assert.Equal(t, "The content of the manifest.", manifestElemSchema.Schema["content"].Description)
	assert.NotNil(t, manifestElemSchema.Schema["content"].DiffSuppressFunc)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["tag"].Type)
	assert.Equal(t, true, elemSchema.Schema["tag"].Optional)
	assert.Equal(t, "The identifier or version to label the pack.", elemSchema.Schema["tag"].Description)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["values"].Type)
	assert.Equal(t, true, elemSchema.Schema["values"].Optional)
	assert.Equal(t, "The values to be used for the pack. This is a stringified JSON object.", elemSchema.Schema["values"].Description)
	assert.NotNil(t, elemSchema.Schema["values"].DiffSuppressFunc)
}

func TestClusterLocationSchema(t *testing.T) {
	s := ClusterLocationSchema()

	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, true, s.Optional)

	assert.NotNil(t, s.DiffSuppressFunc)

	elemSchema, ok := s.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, elemSchema)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["country_code"].Type)
	assert.Equal(t, true, elemSchema.Schema["country_code"].Optional)
	assert.Equal(t, "", elemSchema.Schema["country_code"].Default)
	assert.Equal(t, "The country code of the country the cluster is located in.", elemSchema.Schema["country_code"].Description)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["country_name"].Type)
	assert.Equal(t, true, elemSchema.Schema["country_name"].Optional)
	assert.Equal(t, "", elemSchema.Schema["country_name"].Default)
	assert.Equal(t, "The name of the country.", elemSchema.Schema["country_name"].Description)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["region_code"].Type)
	assert.Equal(t, true, elemSchema.Schema["region_code"].Optional)
	assert.Equal(t, "", elemSchema.Schema["region_code"].Default)
	assert.Equal(t, "The region code of where the cluster is located in.", elemSchema.Schema["region_code"].Description)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["region_name"].Type)
	assert.Equal(t, true, elemSchema.Schema["region_name"].Optional)
	assert.Equal(t, "", elemSchema.Schema["region_name"].Default)
	assert.Equal(t, "The name of the region.", elemSchema.Schema["region_name"].Description)

	assert.Equal(t, schema.TypeFloat, elemSchema.Schema["latitude"].Type)
	assert.Equal(t, true, elemSchema.Schema["latitude"].Required)
	assert.Equal(t, "The latitude coordinates value.", elemSchema.Schema["latitude"].Description)

	assert.Equal(t, schema.TypeFloat, elemSchema.Schema["longitude"].Type)
	assert.Equal(t, true, elemSchema.Schema["longitude"].Required)
	assert.Equal(t, "The longitude coordinates value.", elemSchema.Schema["longitude"].Description)
}

func TestVMVolumeSchema(t *testing.T) {
	s := VMVolumeSchema()

	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, true, s.Optional)

	assert.NotNil(t, s.Elem)
	elemSchema, ok := s.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, elemSchema)

	assert.NotNil(t, elemSchema.Schema["name"])
	assert.Equal(t, schema.TypeString, elemSchema.Schema["name"].Type)
	assert.Equal(t, true, elemSchema.Schema["name"].Required)

	assert.NotNil(t, elemSchema.Schema["container_disk"])
	assert.Equal(t, schema.TypeSet, elemSchema.Schema["container_disk"].Type)
	assert.Equal(t, true, elemSchema.Schema["container_disk"].Optional)
	assert.NotNil(t, elemSchema.Schema["container_disk"].Elem)
	containerDiskSchema, ok := elemSchema.Schema["container_disk"].Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, containerDiskSchema.Schema["image_url"].Type)
	assert.Equal(t, true, containerDiskSchema.Schema["image_url"].Required)

	assert.NotNil(t, elemSchema.Schema["cloud_init_no_cloud"])
	assert.Equal(t, schema.TypeSet, elemSchema.Schema["cloud_init_no_cloud"].Type)
	assert.Equal(t, true, elemSchema.Schema["cloud_init_no_cloud"].Optional)
	assert.NotNil(t, elemSchema.Schema["cloud_init_no_cloud"].Elem)
	cloudInitDiskSchema, ok := elemSchema.Schema["cloud_init_no_cloud"].Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, cloudInitDiskSchema.Schema["user_data"].Type)
	assert.Equal(t, true, cloudInitDiskSchema.Schema["user_data"].Required)

	assert.NotNil(t, elemSchema.Schema["data_volume"])
	assert.Equal(t, schema.TypeSet, elemSchema.Schema["data_volume"].Type)
	assert.Equal(t, true, elemSchema.Schema["data_volume"].Optional)
	assert.NotNil(t, elemSchema.Schema["data_volume"].Elem)
	dataVolumeSchema, ok := elemSchema.Schema["data_volume"].Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, dataVolumeSchema.Schema["storage"].Type)
	assert.Equal(t, true, dataVolumeSchema.Schema["storage"].Required)
}

func TestResourceContainerDiskHash(t *testing.T) {
	v := map[string]interface{}{
		"image_url": "http://example.com/image",
	}
	expected := int(hash("http://example.com/image-"))
	assert.Equal(t, expected, resourceContainerDiskHash(v))
}

func TestResourceCloudInitDiskHash(t *testing.T) {
	v := map[string]interface{}{
		"user_data": "user-data-content",
	}
	expected := int(hash("user-data-content-"))
	assert.Equal(t, expected, resourceCloudInitDiskHash(v))
}

func TestResourceDataVolumeHash(t *testing.T) {
	v := map[string]interface{}{
		"storage": "100GiB",
	}
	expected := int(hash("100GiB-"))
	assert.Equal(t, expected, resourceDataVolumeHash(v))
}

func TestVMNicSchema(t *testing.T) {
	s := VMNicSchema()

	assert.Equal(t, schema.TypeSet, s.Type)
	assert.Equal(t, true, s.Optional)

	assert.NotNil(t, s.Elem)
	elemSchema, ok := s.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, elemSchema)

	assert.NotNil(t, elemSchema.Schema["nic"])
	assert.Equal(t, schema.TypeList, elemSchema.Schema["nic"].Type)
	assert.Equal(t, true, elemSchema.Schema["nic"].Optional)
	assert.NotNil(t, elemSchema.Schema["nic"].Elem)
	nicElemSchema, ok := elemSchema.Schema["nic"].Elem.(*schema.Resource)
	assert.True(t, ok)

	assert.NotNil(t, nicElemSchema.Schema["name"])
	assert.Equal(t, schema.TypeString, nicElemSchema.Schema["name"].Type)
	assert.Equal(t, true, nicElemSchema.Schema["name"].Required)

	assert.NotNil(t, nicElemSchema.Schema["multus"])
	assert.Equal(t, schema.TypeList, nicElemSchema.Schema["multus"].Type)
	assert.Equal(t, true, nicElemSchema.Schema["multus"].Optional)
	assert.Equal(t, 1, nicElemSchema.Schema["multus"].MaxItems)
	assert.NotNil(t, nicElemSchema.Schema["multus"].Elem)
	multusElemSchema, ok := nicElemSchema.Schema["multus"].Elem.(*schema.Resource)
	assert.True(t, ok)

	assert.NotNil(t, multusElemSchema.Schema["network_name"])
	assert.Equal(t, schema.TypeString, multusElemSchema.Schema["network_name"].Type)
	assert.Equal(t, true, multusElemSchema.Schema["network_name"].Required)

	assert.NotNil(t, multusElemSchema.Schema["default"])
	assert.Equal(t, schema.TypeBool, multusElemSchema.Schema["default"].Type)
	assert.Equal(t, true, multusElemSchema.Schema["default"].Optional)

	assert.NotNil(t, nicElemSchema.Schema["network_type"])
	assert.Equal(t, schema.TypeString, nicElemSchema.Schema["network_type"].Type)
	assert.Equal(t, true, nicElemSchema.Schema["network_type"].Optional)
}

func TestVMInterfaceSchema(t *testing.T) {
	s := VMInterfaceSchema()

	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, true, s.Required)

	assert.NotNil(t, s.Elem)
	elemSchema, ok := s.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, elemSchema)

	assert.NotNil(t, elemSchema.Schema["name"])
	assert.Equal(t, schema.TypeString, elemSchema.Schema["name"].Type)
	assert.Equal(t, true, elemSchema.Schema["name"].Required)

	assert.NotNil(t, elemSchema.Schema["type"])
	assert.Equal(t, schema.TypeString, elemSchema.Schema["type"].Type)
	assert.Equal(t, true, elemSchema.Schema["type"].Optional)
	assert.Equal(t, "masquerade", elemSchema.Schema["type"].Default)

	assert.NotNil(t, elemSchema.Schema["model"])
	assert.Equal(t, schema.TypeString, elemSchema.Schema["model"].Type)
	assert.Equal(t, true, elemSchema.Schema["model"].Optional)
	assert.Equal(t, "virtio", elemSchema.Schema["model"].Default)
}

func TestVMDiskSchema(t *testing.T) {
	s := VMDiskSchema()

	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, true, s.Required)

	assert.NotNil(t, s.Elem)
	elemSchema, ok := s.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, elemSchema)

	assert.NotNil(t, elemSchema.Schema["name"])
	assert.Equal(t, schema.TypeString, elemSchema.Schema["name"].Type)
	assert.Equal(t, true, elemSchema.Schema["name"].Required)

	assert.NotNil(t, elemSchema.Schema["bus"])
	assert.Equal(t, schema.TypeString, elemSchema.Schema["bus"].Type)
	assert.Equal(t, true, elemSchema.Schema["bus"].Required)
}

func TestVMDeviceSchema(t *testing.T) {
	s := VMDeviceSchema()

	assert.Equal(t, schema.TypeSet, s.Type)
	assert.Equal(t, true, s.Optional)
	assert.Equal(t, 1, s.MaxItems)

	assert.NotNil(t, s.Elem)
	elemSchema, ok := s.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, elemSchema)

	// Test 'disk' schema
	diskSchema, ok := elemSchema.Schema["disk"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeList, diskSchema.Type)
	assert.Equal(t, true, diskSchema.Required)
	assert.NotNil(t, diskSchema.Elem)

	// Test 'interface' schema
	interfaceSchema, ok := elemSchema.Schema["interface"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeList, interfaceSchema.Type)
	assert.Equal(t, true, interfaceSchema.Required)
	assert.NotNil(t, interfaceSchema.Elem)
}

func TestPackSchema(t *testing.T) {
	s := PackSchema()

	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, true, s.Optional)
	assert.Equal(t, "For packs of type `spectro`, `helm`, and `manifest`, at least one pack must be specified.", s.Description)

	assert.NotNil(t, s.Elem)
	elemSchema, ok := s.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, elemSchema)

	// Test 'uid' schema
	uidSchema, ok := elemSchema.Schema["uid"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, uidSchema.Type)
	assert.Equal(t, true, uidSchema.Optional)
	assert.Equal(t, true, uidSchema.Computed)
	assert.Equal(t, "The unique identifier of the pack. The value can be looked up using the [`spectrocloud_pack`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs/data-sources/pack) data source. This value is required if the pack type is `spectro` and for `helm` if the chart is from a public helm registry. If not provided, all of `name`, `tag`, and `registry_uid` must be specified to resolve the pack UID internally.", uidSchema.Description)

	// Test 'type' schema
	typeSchema, ok := elemSchema.Schema["type"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, typeSchema.Type)
	assert.Equal(t, true, typeSchema.Optional)
	assert.Equal(t, "spectro", typeSchema.Default)
	assert.Equal(t, "The type of the pack. Allowed values are `spectro`, `manifest`, `helm`, or `oci`. The default value is spectro. If using an OCI registry for pack, set the type to `oci`.", typeSchema.Description)

	// Test 'name' schema
	nameSchema, ok := elemSchema.Schema["name"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, nameSchema.Type)
	assert.Equal(t, true, nameSchema.Required)
	assert.Equal(t, "The name of the pack. The name must be unique within the cluster profile. ", nameSchema.Description)

	// Test 'registry_uid' schema
	registryUIDSchema, ok := elemSchema.Schema["registry_uid"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, registryUIDSchema.Type)
	assert.Equal(t, true, registryUIDSchema.Optional)
	assert.Equal(t, "The registry UID of the pack. The registry UID is the unique identifier of the registry. This attribute is required if there is more than one registry that contains a pack with the same name. If `uid` is not provided, this field is required along with `name` and `tag` to resolve the pack UID internally.", registryUIDSchema.Description)

	// Test 'tag' schema
	tagSchema, ok := elemSchema.Schema["tag"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, tagSchema.Type)
	assert.Equal(t, true, tagSchema.Optional)
	assert.Equal(t, "The tag of the pack. The tag is the version of the pack. This attribute is required if the pack type is `spectro` or `helm`. If `uid` is not provided, this field is required along with `name` and `registry_uid` to resolve the pack UID internally.", tagSchema.Description)

	// Test 'values' schema
	valuesSchema, ok := elemSchema.Schema["values"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, valuesSchema.Type)
	assert.Equal(t, true, valuesSchema.Optional)
	assert.Equal(t, "The values of the pack. The values are the configuration values of the pack. The values are specified in YAML format. ", valuesSchema.Description)

	// Test 'manifest' schema
	manifestSchema, ok := elemSchema.Schema["manifest"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeList, manifestSchema.Type)
	assert.Equal(t, true, manifestSchema.Optional)
	assert.NotNil(t, manifestSchema.Elem)
	manifestElemSchema, ok := manifestSchema.Elem.(*schema.Resource)
	assert.True(t, ok)
	assert.NotNil(t, manifestElemSchema)

	// Test 'manifest' nested schema
	manifestUIDSchema, ok := manifestElemSchema.Schema["uid"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, manifestUIDSchema.Type)
	assert.Equal(t, true, manifestUIDSchema.Computed)

	manifestNameSchema, ok := manifestElemSchema.Schema["name"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, manifestNameSchema.Type)
	assert.Equal(t, true, manifestNameSchema.Required)
	assert.Equal(t, "The name of the manifest. The name must be unique within the pack. ", manifestNameSchema.Description)

	manifestContentSchema, ok := manifestElemSchema.Schema["content"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, manifestContentSchema.Type)
	assert.Equal(t, true, manifestContentSchema.Required)
	assert.Equal(t, "The content of the manifest. The content is the YAML content of the manifest. ", manifestContentSchema.Description)
}
