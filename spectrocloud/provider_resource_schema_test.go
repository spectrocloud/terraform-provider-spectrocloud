package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func prepareTeamTestData(id string) *schema.ResourceData {
	d := resourceTeam().TestResourceData()
	d.SetId(id)
	return d
}

func prepareProjectTestData(id string) *schema.ResourceData {
	d := resourceProject().TestResourceData()
	d.SetId(id)
	return d
}

func prepareApplicationProfileTestData(id string) *schema.ResourceData {
	d := resourceApplicationProfile().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterProfileTestData(id string) *schema.ResourceData {
	d := resourceClusterProfile().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterProfileImportTestData(id string) *schema.ResourceData {
	d := resourceClusterProfileImportFeature().TestResourceData()
	d.SetId(id)
	return d
}

func prepareCloudAccountAwsTestData(id string) *schema.ResourceData {
	d := resourceCloudAccountAws().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterAwsTestData(id string) *schema.ResourceData {
	d := resourceClusterAws().TestResourceData()
	d.SetId(id)
	return d
}

func prepareCloudAccountMaasTestData(id string) *schema.ResourceData {
	d := resourceCloudAccountMaas().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterMaasTestData(id string) *schema.ResourceData {
	d := resourceClusterMaas().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterEksTestData(id string) *schema.ResourceData {
	d := resourceClusterEks().TestResourceData()
	d.SetId(id)
	return d
}

func prepareCloudAccountAzureTestData(id string) *schema.ResourceData {
	d := resourceCloudAccountAzure().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterAzureTestData(id string) *schema.ResourceData {
	d := resourceClusterAzure().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterAksTestData(id string) *schema.ResourceData {
	d := resourceClusterAks().TestResourceData()
	d.SetId(id)
	return d
}

func prepareCloudAccountGcpTestData(id string) *schema.ResourceData {
	d := resourceCloudAccountGcp().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterGcpTestData(id string) *schema.ResourceData {
	d := resourceClusterGcp().TestResourceData()
	d.SetId(id)
	return d
}

func prepareCloudAccountOpenstackTestData(id string) *schema.ResourceData {
	d := resourceCloudAccountOpenstack().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterOpenStackTestData(id string) *schema.ResourceData {
	d := resourceClusterOpenStack().TestResourceData()
	d.SetId(id)
	return d
}

func prepareCloudAccountVsphereTestData(id string) *schema.ResourceData {
	d := resourceCloudAccountVsphere().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterEdgeNativeTestData(id string) *schema.ResourceData {
	d := resourceClusterEdgeNative().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterEdgeVsphereTestData(id string) *schema.ResourceData {
	d := resourceClusterEdgeVsphere().TestResourceData()
	d.SetId(id)
	return d
}

func prepareClusterVirtualTestData(id string) *schema.ResourceData {
	d := resourceClusterVirtual().TestResourceData()
	d.SetId(id)
	return d
}

func prepareAddonDeploymentTestData(id string) *schema.ResourceData {
	d := resourceAddonDeployment().TestResourceData()
	d.SetId(id)

	// Set the cluster_uid, cluster_context, and apply_setting fields
	err := d.Set("cluster_uid", "cluster-123")
	if err != nil {
		return nil
	}
	err = d.Set("context", "tenant")
	if err != nil {
		return nil
	}
	err = d.Set("apply_setting", "test-setting")
	if err != nil {
		return nil
	}

	// Set up the cluster_profile field
	profiles := []interface{}{
		map[string]interface{}{
			"id": "profile-1",
			"pack": []interface{}{
				map[string]interface{}{
					"name":   "pack-1",
					"values": "value-1",
					"tag":    "tag-1",
					"type":   "type-1",
					"manifest": []interface{}{
						map[string]interface{}{
							"name":    "manifest-1",
							"content": "content-1",
						},
					},
				},
			},
		},
		map[string]interface{}{
			"id": "profile-2",
			"pack": []interface{}{
				map[string]interface{}{
					"name":   "pack-2",
					"values": "value-2",
					"tag":    "tag-2",
					"type":   "type-2",
					"manifest": []interface{}{
						map[string]interface{}{
							"name":    "manifest-2",
							"content": "content-2",
						},
					},
				},
			},
		},
	}
	err = d.Set("cluster_profile", profiles)
	if err != nil {
		return nil
	}

	return d
}

func prepareKubevirtVirtualMachineTestData(id string) *schema.ResourceData {
	d := resourceKubevirtVirtualMachine().TestResourceData()
	d.SetId(id)
	return d
}

func prepareKubevirtDataVolumeTestData(id string) *schema.ResourceData {
	d := resourceKubevirtDataVolume().TestResourceData()
	d.SetId(id)
	return d
}

func preparePrivateCloudGatewayIpPoolTestData(id string) *schema.ResourceData {
	d := resourcePrivateCloudGatewayIpPool().TestResourceData()
	d.SetId(id)
	return d
}

func prepareBackupStorageLocationTestData(id string) *schema.ResourceData {
	d := resourceBackupStorageLocation().TestResourceData()
	d.SetId(id)
	return d
}

func prepareRegistryOciEcrTestData(id string) *schema.ResourceData {
	d := resourceRegistryOciEcr().TestResourceData()
	d.SetId(id)
	return d
}

func prepareRegistryHelmTestData(id string) *schema.ResourceData {
	d := resourceRegistryHelm().TestResourceData()
	d.SetId(id)
	return d
}

func prepareApplianceTestData(id string) *schema.ResourceData {
	d := resourceAppliance().TestResourceData()
	d.SetId(id)
	return d
}

func prepareWorkspaceTestData(id string) *schema.ResourceData {
	d := resourceWorkspace().TestResourceData()
	d.SetId(id)
	return d
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////
func TestResourceTeam(t *testing.T) {
	testData := prepareTeamTestData("test-id")
	// assert id is the same
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceProject(t *testing.T) {
	testData := prepareProjectTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceApplicationProfile(t *testing.T) {
	testData := prepareApplicationProfileTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterProfile(t *testing.T) {
	testData := prepareClusterProfileTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterProfileImport(t *testing.T) {
	testData := prepareClusterProfileImportTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceCloudAccountAws(t *testing.T) {
	testData := prepareCloudAccountAwsTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterAws(t *testing.T) {
	testData := prepareClusterAwsTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceCloudAccountMaas(t *testing.T) {
	testData := prepareCloudAccountMaasTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterMaas(t *testing.T) {
	testData := prepareClusterMaasTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterEks(t *testing.T) {
	testData := prepareClusterEksTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceCloudAccountAzure(t *testing.T) {
	testData := prepareCloudAccountAzureTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterAzure(t *testing.T) {
	testData := prepareClusterAzureTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterAks(t *testing.T) {
	testData := prepareClusterAksTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceCloudAccountGcp(t *testing.T) {
	testData := prepareCloudAccountGcpTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterGcp(t *testing.T) {
	testData := prepareClusterGcpTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceCloudAccountOpenstack(t *testing.T) {
	testData := prepareCloudAccountOpenstackTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterOpenStack(t *testing.T) {
	testData := prepareClusterOpenStackTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceCloudAccountVsphere(t *testing.T) {
	testData := prepareCloudAccountVsphereTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterEdgeNative(t *testing.T) {
	testData := prepareClusterEdgeNativeTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterEdgeVsphere(t *testing.T) {
	testData := prepareClusterEdgeVsphereTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceClusterVirtual(t *testing.T) {
	testData := prepareClusterVirtualTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceAddonDeployment(t *testing.T) {
	testData := prepareAddonDeploymentTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceKubevirtVirtualMachine(t *testing.T) {
	testData := prepareKubevirtVirtualMachineTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceKubevirtDataVolume(t *testing.T) {
	testData := prepareKubevirtDataVolumeTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourcePrivateCloudGatewayIpPool(t *testing.T) {
	testData := preparePrivateCloudGatewayIpPoolTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceBackupStorageLocation(t *testing.T) {
	testData := prepareBackupStorageLocationTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceRegistryOciEcr(t *testing.T) {
	testData := prepareRegistryOciEcrTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceRegistryHelm(t *testing.T) {
	testData := prepareRegistryHelmTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceAppliance(t *testing.T) {
	testData := prepareApplianceTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}

func TestResourceWorkspace(t *testing.T) {
	testData := prepareWorkspaceTestData("test-id")
	assert.Equal(t, "test-id", testData.Id())
}
