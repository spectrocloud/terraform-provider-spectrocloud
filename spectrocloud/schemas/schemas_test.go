package schemas

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "The unique id of the registry to be used for the pack. Either `registry_uid` or `registry_name` can be specified, but not both.", elemSchema.Schema["registry_uid"].Description)

	assert.Equal(t, schema.TypeString, elemSchema.Schema["registry_name"].Type)
	assert.Equal(t, true, elemSchema.Schema["registry_name"].Optional)
	assert.Equal(t, "The name of the registry to be used for the pack. This can be used instead of `registry_uid` for better readability. Either `registry_uid` or `registry_name` can be specified, but not both.", elemSchema.Schema["registry_name"].Description)

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
	assert.Equal(t, "The registry UID of the pack. The registry UID is the unique identifier of the registry. This attribute is required if there is more than one registry that contains a pack with the same name. If `uid` is not provided, this field is required along with `name` and `tag` to resolve the pack UID internally. Either `registry_uid` or `registry_name` can be specified, but not both.", registryUIDSchema.Description)

	// Test 'registry_name' schema
	registryNameSchema, ok := elemSchema.Schema["registry_name"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, registryNameSchema.Type)
	assert.Equal(t, true, registryNameSchema.Optional)
	assert.Equal(t, "The registry name of the pack. The registry name is the human-readable name of the registry. This attribute can be used instead of `registry_uid` for better readability. If `uid` is not provided, this field can be used along with `name` and `tag` to resolve the pack UID internally. Either `registry_uid` or `registry_name` can be specified, but not both.", registryNameSchema.Description)

	// Test 'tag' schema
	tagSchema, ok := elemSchema.Schema["tag"]
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, tagSchema.Type)
	assert.Equal(t, true, tagSchema.Optional)
	assert.Equal(t, "The tag of the pack. The tag is the version of the pack. This attribute is required if the pack type is `spectro` or `helm`. If `uid` is not provided, this field is required along with `name` and `registry_uid` (or `registry_name`) to resolve the pack UID internally.", tagSchema.Description)

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

func TestSubnetSchema(t *testing.T) {
	subnetSchema := SubnetSchema()

	// Test root schema properties
	assert.Equal(t, schema.TypeList, subnetSchema.Type, "Subnet schema should be TypeList")
	assert.True(t, subnetSchema.Optional, "Subnet schema should be optional")
	assert.Equal(t, 1, subnetSchema.MaxItems, "Subnet schema should have MaxItems of 1")
	assert.NotNil(t, subnetSchema.RequiredWith, "Subnet schema should have RequiredWith")
	assert.Contains(t, subnetSchema.RequiredWith, "cloud_config.0.network_resource_group", "RequiredWith should include network_resource_group")
	assert.Contains(t, subnetSchema.RequiredWith, "cloud_config.0.virtual_network_name", "RequiredWith should include virtual_network_name")
	assert.Contains(t, subnetSchema.RequiredWith, "cloud_config.0.virtual_network_cidr_block", "RequiredWith should include virtual_network_cidr_block")

	// Test that Elem is a Resource
	elemResource, ok := subnetSchema.Elem.(*schema.Resource)
	assert.True(t, ok, "Subnet schema Elem should be a Resource")
	assert.NotNil(t, elemResource, "Subnet schema Elem Resource should not be nil")

	// Test nested schema fields
	nestedSchema := elemResource.Schema
	assert.NotNil(t, nestedSchema, "Nested schema should not be nil")

	// Test "name" field
	nameSchema, exists := nestedSchema["name"]
	assert.True(t, exists, "Nested schema should have 'name' field")
	assert.NotNil(t, nameSchema, "'name' field schema should not be nil")
	assert.Equal(t, schema.TypeString, nameSchema.Type, "'name' field should be TypeString")
	assert.True(t, nameSchema.Required, "'name' field should be required")
	assert.False(t, nameSchema.Optional, "'name' field should not be optional")
	assert.Equal(t, "Name of the subnet.", nameSchema.Description, "'name' field should have correct description")

	// Test "cidr_block" field
	cidrBlockSchema, exists := nestedSchema["cidr_block"]
	assert.True(t, exists, "Nested schema should have 'cidr_block' field")
	assert.NotNil(t, cidrBlockSchema, "'cidr_block' field schema should not be nil")
	assert.Equal(t, schema.TypeString, cidrBlockSchema.Type, "'cidr_block' field should be TypeString")
	assert.True(t, cidrBlockSchema.Required, "'cidr_block' field should be required")
	assert.False(t, cidrBlockSchema.Optional, "'cidr_block' field should not be optional")
	assert.Equal(t, "CidrBlock is the CIDR block to be used when the provider creates a managed virtual network.", cidrBlockSchema.Description, "'cidr_block' field should have correct description")

	// Test "security_group_name" field
	securityGroupNameSchema, exists := nestedSchema["security_group_name"]
	assert.True(t, exists, "Nested schema should have 'security_group_name' field")
	assert.NotNil(t, securityGroupNameSchema, "'security_group_name' field schema should not be nil")
	assert.Equal(t, schema.TypeString, securityGroupNameSchema.Type, "'security_group_name' field should be TypeString")
	assert.True(t, securityGroupNameSchema.Optional, "'security_group_name' field should be optional")
	assert.False(t, securityGroupNameSchema.Required, "'security_group_name' field should not be required")
	assert.Equal(t, "Network Security Group(NSG) to be attached to subnet.", securityGroupNameSchema.Description, "'security_group_name' field should have correct description")
}

func TestScanPolicySchema(t *testing.T) {
	scanPolicySchema := ScanPolicySchema()

	// Test root schema properties
	assert.Equal(t, schema.TypeList, scanPolicySchema.Type, "ScanPolicy schema should be TypeList")
	assert.True(t, scanPolicySchema.Optional, "ScanPolicy schema should be optional")
	assert.Equal(t, 1, scanPolicySchema.MaxItems, "ScanPolicy schema should have MaxItems of 1")
	assert.Equal(t, "The scan policy for the cluster.", scanPolicySchema.Description, "ScanPolicy schema should have correct description")

	// Test that Elem is a Resource
	elemResource, ok := scanPolicySchema.Elem.(*schema.Resource)
	assert.True(t, ok, "ScanPolicy schema Elem should be a Resource")
	assert.NotNil(t, elemResource, "ScanPolicy schema Elem Resource should not be nil")

	// Test nested schema fields
	nestedSchema := elemResource.Schema
	assert.NotNil(t, nestedSchema, "Nested schema should not be nil")

	// Test "configuration_scan_schedule" field
	configScanSchema, exists := nestedSchema["configuration_scan_schedule"]
	assert.True(t, exists, "Nested schema should have 'configuration_scan_schedule' field")
	assert.NotNil(t, configScanSchema, "'configuration_scan_schedule' field schema should not be nil")
	assert.Equal(t, schema.TypeString, configScanSchema.Type, "'configuration_scan_schedule' field should be TypeString")
	assert.True(t, configScanSchema.Required, "'configuration_scan_schedule' field should be required")
	assert.False(t, configScanSchema.Optional, "'configuration_scan_schedule' field should not be optional")
	assert.Equal(t, "The schedule for configuration scan.", configScanSchema.Description, "'configuration_scan_schedule' field should have correct description")

	// Test "penetration_scan_schedule" field
	penetrationScanSchema, exists := nestedSchema["penetration_scan_schedule"]
	assert.True(t, exists, "Nested schema should have 'penetration_scan_schedule' field")
	assert.NotNil(t, penetrationScanSchema, "'penetration_scan_schedule' field schema should not be nil")
	assert.Equal(t, schema.TypeString, penetrationScanSchema.Type, "'penetration_scan_schedule' field should be TypeString")
	assert.True(t, penetrationScanSchema.Required, "'penetration_scan_schedule' field should be required")
	assert.False(t, penetrationScanSchema.Optional, "'penetration_scan_schedule' field should not be optional")
	assert.Equal(t, "The schedule for penetration scan.", penetrationScanSchema.Description, "'penetration_scan_schedule' field should have correct description")

	// Test "conformance_scan_schedule" field
	conformanceScanSchema, exists := nestedSchema["conformance_scan_schedule"]
	assert.True(t, exists, "Nested schema should have 'conformance_scan_schedule' field")
	assert.NotNil(t, conformanceScanSchema, "'conformance_scan_schedule' field schema should not be nil")
	assert.Equal(t, schema.TypeString, conformanceScanSchema.Type, "'conformance_scan_schedule' field should be TypeString")
	assert.True(t, conformanceScanSchema.Required, "'conformance_scan_schedule' field should be required")
	assert.False(t, conformanceScanSchema.Optional, "'conformance_scan_schedule' field should not be optional")
	assert.Equal(t, "The schedule for conformance scan.", conformanceScanSchema.Description, "'conformance_scan_schedule' field should have correct description")
}

func TestProfileVariables(t *testing.T) {
	profileVarsSchema := ProfileVariables()

	// Test root schema properties
	assert.Equal(t, schema.TypeList, profileVarsSchema.Type, "ProfileVariables schema should be TypeList")
	assert.True(t, profileVarsSchema.Optional, "ProfileVariables schema should be optional")
	assert.Equal(t, 1, profileVarsSchema.MaxItems, "ProfileVariables schema should have MaxItems of 1")
	assert.Equal(t, "List of variables for the cluster profile.", profileVarsSchema.Description, "ProfileVariables schema should have correct description")

	// Test that Elem is a Resource
	elemResource, ok := profileVarsSchema.Elem.(*schema.Resource)
	assert.True(t, ok, "ProfileVariables schema Elem should be a Resource")
	assert.NotNil(t, elemResource, "ProfileVariables schema Elem Resource should not be nil")

	// Test first level nested schema - should have "variable" field
	firstLevelSchema := elemResource.Schema
	assert.NotNil(t, firstLevelSchema, "First level nested schema should not be nil")
	assert.Equal(t, 1, len(firstLevelSchema), "First level schema should have exactly 1 field")

	// Test "variable" field
	variableSchema, exists := firstLevelSchema["variable"]
	assert.True(t, exists, "First level schema should have 'variable' field")
	assert.NotNil(t, variableSchema, "'variable' field schema should not be nil")
	assert.Equal(t, schema.TypeList, variableSchema.Type, "'variable' field should be TypeList")
	assert.True(t, variableSchema.Required, "'variable' field should be required")
	assert.False(t, variableSchema.Optional, "'variable' field should not be optional")

	// Test that variable Elem is a Resource
	variableElemResource, ok := variableSchema.Elem.(*schema.Resource)
	assert.True(t, ok, "'variable' field Elem should be a Resource")
	assert.NotNil(t, variableElemResource, "'variable' field Elem Resource should not be nil")

	// Test variable Resource schema fields
	variableResourceSchema := variableElemResource.Schema
	assert.NotNil(t, variableResourceSchema, "Variable Resource schema should not be nil")
	assert.Equal(t, 10, len(variableResourceSchema), "Variable Resource schema should have exactly 10 fields")

	// Test "name" field
	nameSchema, exists := variableResourceSchema["name"]
	assert.True(t, exists, "Variable schema should have 'name' field")
	assert.NotNil(t, nameSchema, "'name' field schema should not be nil")
	assert.Equal(t, schema.TypeString, nameSchema.Type, "'name' field should be TypeString")
	assert.True(t, nameSchema.Required, "'name' field should be required")
	assert.False(t, nameSchema.Optional, "'name' field should not be optional")
	assert.Equal(t, "The name of the variable should be unique among variables.", nameSchema.Description, "'name' field should have correct description")

	// Test "display_name" field
	displayNameSchema, exists := variableResourceSchema["display_name"]
	assert.True(t, exists, "Variable schema should have 'display_name' field")
	assert.NotNil(t, displayNameSchema, "'display_name' field schema should not be nil")
	assert.Equal(t, schema.TypeString, displayNameSchema.Type, "'display_name' field should be TypeString")
	assert.True(t, displayNameSchema.Required, "'display_name' field should be required")
	assert.False(t, displayNameSchema.Optional, "'display_name' field should not be optional")
	assert.Equal(t, "The display name of the variable should be unique among variables.", displayNameSchema.Description, "'display_name' field should have correct description")

	// Test "format" field
	formatSchema, exists := variableResourceSchema["format"]
	assert.True(t, exists, "Variable schema should have 'format' field")
	assert.NotNil(t, formatSchema, "'format' field schema should not be nil")
	assert.Equal(t, schema.TypeString, formatSchema.Type, "'format' field should be TypeString")
	assert.True(t, formatSchema.Optional, "'format' field should be optional")
	assert.False(t, formatSchema.Required, "'format' field should not be required")
	assert.Equal(t, "string", formatSchema.Default, "'format' field should have default value 'string'")
	assert.NotNil(t, formatSchema.ValidateFunc, "'format' field should have ValidateFunc")
	assert.Equal(t, "The format of the variable. Default is `string`, `format` field can only be set during cluster profile creation. Allowed formats include `string`, `number`, `boolean`, `ipv4`, `ipv4cidr`, `ipv6`, `version`.", formatSchema.Description, "'format' field should have correct description")

	// Test "description" field
	descriptionSchema, exists := variableResourceSchema["description"]
	assert.True(t, exists, "Variable schema should have 'description' field")
	assert.NotNil(t, descriptionSchema, "'description' field schema should not be nil")
	assert.Equal(t, schema.TypeString, descriptionSchema.Type, "'description' field should be TypeString")
	assert.True(t, descriptionSchema.Optional, "'description' field should be optional")
	assert.False(t, descriptionSchema.Required, "'description' field should not be required")
	assert.Equal(t, "The description of the variable.", descriptionSchema.Description, "'description' field should have correct description")

	// Test "default_value" field
	defaultValueSchema, exists := variableResourceSchema["default_value"]
	assert.True(t, exists, "Variable schema should have 'default_value' field")
	assert.NotNil(t, defaultValueSchema, "'default_value' field schema should not be nil")
	assert.Equal(t, schema.TypeString, defaultValueSchema.Type, "'default_value' field should be TypeString")
	assert.True(t, defaultValueSchema.Optional, "'default_value' field should be optional")
	assert.False(t, defaultValueSchema.Required, "'default_value' field should not be required")
	assert.Equal(t, "The default value of the variable.", defaultValueSchema.Description, "'default_value' field should have correct description")

	// Test "regex" field
	regexSchema, exists := variableResourceSchema["regex"]
	assert.True(t, exists, "Variable schema should have 'regex' field")
	assert.NotNil(t, regexSchema, "'regex' field schema should not be nil")
	assert.Equal(t, schema.TypeString, regexSchema.Type, "'regex' field should be TypeString")
	assert.True(t, regexSchema.Optional, "'regex' field should be optional")
	assert.False(t, regexSchema.Required, "'regex' field should not be required")
	assert.Equal(t, "Regular expression pattern which the variable value must match.", regexSchema.Description, "'regex' field should have correct description")

	// Test "required" field
	requiredSchema, exists := variableResourceSchema["required"]
	assert.True(t, exists, "Variable schema should have 'required' field")
	assert.NotNil(t, requiredSchema, "'required' field schema should not be nil")
	assert.Equal(t, schema.TypeBool, requiredSchema.Type, "'required' field should be TypeBool")
	assert.True(t, requiredSchema.Optional, "'required' field should be optional")
	assert.False(t, requiredSchema.Required, "'required' field should not be required")
	assert.Equal(t, "The `required` to specify if the variable is optional or mandatory. If it is mandatory then default value must be provided.", requiredSchema.Description, "'required' field should have correct description")

	// Test "immutable" field
	immutableSchema, exists := variableResourceSchema["immutable"]
	assert.True(t, exists, "Variable schema should have 'immutable' field")
	assert.NotNil(t, immutableSchema, "'immutable' field schema should not be nil")
	assert.Equal(t, schema.TypeBool, immutableSchema.Type, "'immutable' field should be TypeBool")
	assert.True(t, immutableSchema.Optional, "'immutable' field should be optional")
	assert.False(t, immutableSchema.Required, "'immutable' field should not be required")
	assert.Equal(t, "If `immutable` is set to `true`, then variable value can't be editable. By default the `immutable` flag will be set to `false`.", immutableSchema.Description, "'immutable' field should have correct description")

	// Test "is_sensitive" field
	isSensitiveSchema, exists := variableResourceSchema["is_sensitive"]
	assert.True(t, exists, "Variable schema should have 'is_sensitive' field")
	assert.NotNil(t, isSensitiveSchema, "'is_sensitive' field schema should not be nil")
	assert.Equal(t, schema.TypeBool, isSensitiveSchema.Type, "'is_sensitive' field should be TypeBool")
	assert.True(t, isSensitiveSchema.Optional, "'is_sensitive' field should be optional")
	assert.False(t, isSensitiveSchema.Required, "'is_sensitive' field should not be required")
	assert.Equal(t, "If `is_sensitive` is set to `true`, then default value will be masked. By default the `is_sensitive` flag will be set to false.", isSensitiveSchema.Description, "'is_sensitive' field should have correct description")

	// Test "hidden" field
	hiddenSchema, exists := variableResourceSchema["hidden"]
	assert.True(t, exists, "Variable schema should have 'hidden' field")
	assert.NotNil(t, hiddenSchema, "'hidden' field schema should not be nil")
	assert.Equal(t, schema.TypeBool, hiddenSchema.Type, "'hidden' field should be TypeBool")
	assert.True(t, hiddenSchema.Optional, "'hidden' field should be optional")
	assert.False(t, hiddenSchema.Required, "'hidden' field should not be required")
	assert.Equal(t, "If `hidden` is set to `true`, then variable will be hidden for overriding the value. By default the `hidden` flag will be set to `false`.", hiddenSchema.Description, "'hidden' field should have correct description")
}

func TestOverrideScalingSchema(t *testing.T) {
	overrideScalingSchema := OverrideScalingSchema()

	// Test root schema properties
	assert.Equal(t, schema.TypeList, overrideScalingSchema.Type, "OverrideScaling schema should be TypeList")
	assert.True(t, overrideScalingSchema.Optional, "OverrideScaling schema should be optional")
	assert.Equal(t, 1, overrideScalingSchema.MaxItems, "OverrideScaling schema should have MaxItems of 1")
	assert.Equal(t, "Rolling update strategy for the machine pool.", overrideScalingSchema.Description, "OverrideScaling schema should have correct description")

	// Test that Elem is a Resource
	elemResource, ok := overrideScalingSchema.Elem.(*schema.Resource)
	assert.True(t, ok, "OverrideScaling schema Elem should be a Resource")
	assert.NotNil(t, elemResource, "OverrideScaling schema Elem Resource should not be nil")

	// Test nested schema fields
	nestedSchema := elemResource.Schema
	assert.NotNil(t, nestedSchema, "Nested schema should not be nil")

	// Test "max_surge" field
	maxSurgeSchema, exists := nestedSchema["max_surge"]
	assert.True(t, exists, "Nested schema should have 'max_surge' field")
	assert.NotNil(t, maxSurgeSchema, "'max_surge' field schema should not be nil")
	assert.Equal(t, schema.TypeString, maxSurgeSchema.Type, "'max_surge' field should be TypeString")
	assert.True(t, maxSurgeSchema.Optional, "'max_surge' field should be optional")
	assert.False(t, maxSurgeSchema.Required, "'max_surge' field should not be required")
	assert.Equal(t, "", maxSurgeSchema.Default, "'max_surge' field should have default value of empty string")
	assert.Equal(t, "Max extra nodes during rolling update. Integer or percentage (e.g., '1' or '20%'). Only valid when type=OverrideScaling. Both maxSurge and maxUnavailable are required.", maxSurgeSchema.Description, "'max_surge' field should have correct description")

	// Test "max_unavailable" field
	maxUnavailableSchema, exists := nestedSchema["max_unavailable"]
	assert.True(t, exists, "Nested schema should have 'max_unavailable' field")
	assert.NotNil(t, maxUnavailableSchema, "'max_unavailable' field schema should not be nil")
	assert.Equal(t, schema.TypeString, maxUnavailableSchema.Type, "'max_unavailable' field should be TypeString")
	assert.True(t, maxUnavailableSchema.Optional, "'max_unavailable' field should be optional")
	assert.False(t, maxUnavailableSchema.Required, "'max_unavailable' field should not be required")
	assert.Equal(t, "", maxUnavailableSchema.Default, "'max_unavailable' field should have default value of empty string")
	assert.Equal(t, "Max unavailable nodes during rolling update. Integer or percentage (e.g., '0' or '10%'). Only valid when type=OverrideScaling. Both maxSurge and maxUnavailable are required.", maxUnavailableSchema.Description, "'max_unavailable' field should have correct description")
}

func TestAwsLaunchTemplate(t *testing.T) {
	awsLaunchTemplateSchema := AwsLaunchTemplate()

	// Test root schema properties
	assert.Equal(t, schema.TypeList, awsLaunchTemplateSchema.Type, "AwsLaunchTemplate schema should be TypeList")
	assert.True(t, awsLaunchTemplateSchema.Optional, "AwsLaunchTemplate schema should be optional")
	assert.Equal(t, 1, awsLaunchTemplateSchema.MaxItems, "AwsLaunchTemplate schema should have MaxItems of 1")

	// Test that Elem is a Resource
	elemResource, ok := awsLaunchTemplateSchema.Elem.(*schema.Resource)
	assert.True(t, ok, "AwsLaunchTemplate schema Elem should be a Resource")
	assert.NotNil(t, elemResource, "AwsLaunchTemplate schema Elem Resource should not be nil")

	// Test nested schema fields
	nestedSchema := elemResource.Schema
	assert.NotNil(t, nestedSchema, "Nested schema should not be nil")

	// Test "ami_id" field
	amiIdSchema, exists := nestedSchema["ami_id"]
	assert.True(t, exists, "Nested schema should have 'ami_id' field")
	assert.NotNil(t, amiIdSchema, "'ami_id' field schema should not be nil")
	assert.Equal(t, schema.TypeString, amiIdSchema.Type, "'ami_id' field should be TypeString")
	assert.True(t, amiIdSchema.Optional, "'ami_id' field should be optional")
	assert.False(t, amiIdSchema.Required, "'ami_id' field should not be required")
	assert.Equal(t, "The ID of the custom Amazon Machine Image (AMI). If you do not set an `ami_id`, Palette will repave the cluster when it automatically updates the EKS AMI.", amiIdSchema.Description, "'ami_id' field should have correct description")

	// Test "root_volume_type" field
	rootVolumeTypeSchema, exists := nestedSchema["root_volume_type"]
	assert.True(t, exists, "Nested schema should have 'root_volume_type' field")
	assert.NotNil(t, rootVolumeTypeSchema, "'root_volume_type' field schema should not be nil")
	assert.Equal(t, schema.TypeString, rootVolumeTypeSchema.Type, "'root_volume_type' field should be TypeString")
	assert.True(t, rootVolumeTypeSchema.Optional, "'root_volume_type' field should be optional")
	assert.False(t, rootVolumeTypeSchema.Required, "'root_volume_type' field should not be required")
	assert.Equal(t, "The type of the root volume.", rootVolumeTypeSchema.Description, "'root_volume_type' field should have correct description")

	// Test "root_volume_iops" field
	rootVolumeIopsSchema, exists := nestedSchema["root_volume_iops"]
	assert.True(t, exists, "Nested schema should have 'root_volume_iops' field")
	assert.NotNil(t, rootVolumeIopsSchema, "'root_volume_iops' field schema should not be nil")
	assert.Equal(t, schema.TypeInt, rootVolumeIopsSchema.Type, "'root_volume_iops' field should be TypeInt")
	assert.True(t, rootVolumeIopsSchema.Optional, "'root_volume_iops' field should be optional")
	assert.False(t, rootVolumeIopsSchema.Required, "'root_volume_iops' field should not be required")
	assert.Equal(t, "The number of input/output operations per second (IOPS) for the root volume.", rootVolumeIopsSchema.Description, "'root_volume_iops' field should have correct description")

	// Test "root_volume_throughput" field
	rootVolumeThroughputSchema, exists := nestedSchema["root_volume_throughput"]
	assert.True(t, exists, "Nested schema should have 'root_volume_throughput' field")
	assert.NotNil(t, rootVolumeThroughputSchema, "'root_volume_throughput' field schema should not be nil")
	assert.Equal(t, schema.TypeInt, rootVolumeThroughputSchema.Type, "'root_volume_throughput' field should be TypeInt")
	assert.True(t, rootVolumeThroughputSchema.Optional, "'root_volume_throughput' field should be optional")
	assert.False(t, rootVolumeThroughputSchema.Required, "'root_volume_throughput' field should not be required")
	assert.Equal(t, "The throughput of the root volume in MiB/s.", rootVolumeThroughputSchema.Description, "'root_volume_throughput' field should have correct description")

	// Test "additional_security_groups" field
	additionalSecurityGroupsSchema, exists := nestedSchema["additional_security_groups"]
	assert.True(t, exists, "Nested schema should have 'additional_security_groups' field")
	assert.NotNil(t, additionalSecurityGroupsSchema, "'additional_security_groups' field schema should not be nil")
	assert.Equal(t, schema.TypeSet, additionalSecurityGroupsSchema.Type, "'additional_security_groups' field should be TypeSet")
	assert.True(t, additionalSecurityGroupsSchema.Optional, "'additional_security_groups' field should be optional")
	assert.False(t, additionalSecurityGroupsSchema.Required, "'additional_security_groups' field should not be required")
	assert.NotNil(t, additionalSecurityGroupsSchema.Set, "'additional_security_groups' field should have Set function")
	// Note: Cannot directly compare function pointers in Go, but we verify Set is not nil above
	// The actual function assignment (schema.HashString) is verified by the schema definition in eks_template.go
	assert.NotNil(t, additionalSecurityGroupsSchema.Elem, "'additional_security_groups' field should have Elem")
	elemSchema, ok := additionalSecurityGroupsSchema.Elem.(*schema.Schema)
	assert.True(t, ok, "'additional_security_groups' Elem should be a Schema")
	assert.Equal(t, schema.TypeString, elemSchema.Type, "'additional_security_groups' Elem should be TypeString")
	assert.Equal(t, "Additional security groups to attach to the instance.", additionalSecurityGroupsSchema.Description, "'additional_security_groups' field should have correct description")
}
