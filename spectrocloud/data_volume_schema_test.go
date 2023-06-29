package spectrocloud

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func prepareDataVolumeTestData() *schema.ResourceData {
	rd := resourceKubevirtDataVolume().TestResourceData()

	rd.Set("cluster_uid", "cluster-123")
	rd.Set("cluster_context", "tenant")
	rd.Set("vm_name", "vm-test")
	rd.Set("vm_namespace", "default")
	rd.Set("metadata", []interface{}{
		map[string]interface{}{
			"name":      "vol-test",
			"namespace": "default",
			"labels": map[string]interface{}{
				"key1": "value1",
			},
		},
	})
	rd.Set("add_volume_options", []interface{}{
		map[string]interface{}{
			"name": "vol-test",
			"disk": []interface{}{
				map[string]interface{}{
					"name": "vol-test",
					"bus":  "scsi",
				},
			},
			"volume_source": []interface{}{
				map[string]interface{}{
					"data_volume": []interface{}{
						map[string]interface{}{
							"name":         "vol-test",
							"hotpluggable": true,
						},
					},
				},
			},
		},
	})
	rd.Set("spec", []interface{}{
		map[string]interface{}{
			"source": []interface{}{
				map[string]interface{}{
					"http": []interface{}{
						map[string]interface{}{
							"url": "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2",
						},
					},
				},
			},
			"pvc": []interface{}{
				map[string]interface{}{
					"access_modes": []interface{}{
						"ReadWriteOnce",
					},
					"resources": []interface{}{
						map[string]interface{}{
							"requests": map[string]interface{}{
								"storage": "10Gi",
							},
						},
					},
					"storage_class_name": "local-storage",
				},
			},
		},
	})

	return rd
}

func TestCreateDataVolumePositive(t *testing.T) {
	assert := assert.New(t)
	rd := prepareDataVolumeTestData()

	// Mock the V1Client
	m := &client.V1Client{
		GetClusterFn: func(scope string, uid string) (*models.V1SpectroCluster, error) {
			isHost := new(bool)
			*isHost = true
			cluster := &models.V1SpectroCluster{
				APIVersion: "v1",
				Metadata: &models.V1ObjectMeta{
					Annotations:       nil,
					CreationTimestamp: models.V1Time{},
					DeletionTimestamp: models.V1Time{},
					Labels: map[string]string{
						"owner": "siva",
					},
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-vsphere-cluster-unit-test",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "vsphere-uid",
				},
				Spec: &models.V1SpectroClusterSpec{
					CloudConfigRef: &models.V1ObjectReference{
						APIVersion:      "",
						FieldPath:       "",
						Kind:            "",
						Name:            "",
						Namespace:       "",
						ResourceVersion: "",
						UID:             "test-cloud-config-uid",
					},
					CloudType: "",
					ClusterConfig: &models.V1ClusterConfig{
						ClusterRbac:                    nil,
						ClusterResources:               nil,
						ControlPlaneHealthCheckTimeout: "",
						Fips:                           nil,
						HostClusterConfig: &models.V1HostClusterConfig{
							ClusterEndpoint: &models.V1HostClusterEndpoint{
								Config: nil,
								Type:   "LoadBalancer",
							},
							ClusterGroup:  nil,
							HostCluster:   nil,
							IsHostCluster: isHost,
						},
						LifecycleConfig:             nil,
						MachineHealthConfig:         nil,
						MachineManagementConfig:     nil,
						UpdateWorkerPoolsInParallel: false,
					},
					ClusterProfileTemplates: nil,
					ClusterType:             "",
				},
				Status: &models.V1SpectroClusterStatus{
					State: "running",
				},
			}
			return cluster, nil
		},
		CreateDataVolumeFn: func(uid string, name string, body *models.V1VMAddVolumeEntity) (string, error) {
			// Check if input parameters match the expected values
			assert.Equal(uid, "cluster-123")
			assert.Equal(name, "vm-test")
			assert.NotNil(body)

			return "data-volume-id", nil
		},
	}

	ctx := context.Background()
	diags := resourceKubevirtDataVolumeCreate(ctx, rd, m)
	if diags.HasError() {
		assert.Error(errors.New("create operation failed"))
	} else {
		assert.NoError(nil)
	}

	// Check if resourceData ID was set correctly
	expectedID := utils.BuildIdDV("tenant", "cluster-123", "default", "vm-test", &models.V1VMObjectMeta{
		Name:      "vol-test",
		Namespace: "default",
	})
	assert.Equal(expectedID, rd.Id())
}

func TestCreateDataVolume(t *testing.T) {
	rd := prepareDataVolumeTestData()

	m := &client.V1Client{
		CreateDataVolumeFn: func(uid string, name string, body *models.V1VMAddVolumeEntity) (string, error) {
			if uid != "cluster-123" {
				return "", errors.New("unexpected cluster_uid")
			}
			if name != "vm-test" {
				return "", errors.New("unexpected vm_name")
			}
			if body.DataVolumeTemplate.Metadata.Namespace != "default" {
				return "", errors.New("unexpected vm_namespace")
			}
			return "data-volume-id", nil
		},
		GetClusterFn: func(scope string, uid string) (*models.V1SpectroCluster, error) {
			isHost := new(bool)
			*isHost = true
			cluster := &models.V1SpectroCluster{
				APIVersion: "v1",
				Metadata: &models.V1ObjectMeta{
					Annotations:       nil,
					CreationTimestamp: models.V1Time{},
					DeletionTimestamp: models.V1Time{},
					Labels: map[string]string{
						"owner": "siva",
					},
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-vsphere-cluster-unit-test",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "vsphere-uid",
				},
				Spec: &models.V1SpectroClusterSpec{
					CloudConfigRef: &models.V1ObjectReference{
						APIVersion:      "",
						FieldPath:       "",
						Kind:            "",
						Name:            "",
						Namespace:       "",
						ResourceVersion: "",
						UID:             "test-cloud-config-uid",
					},
					CloudType: "",
					ClusterConfig: &models.V1ClusterConfig{
						ClusterRbac:                    nil,
						ClusterResources:               nil,
						ControlPlaneHealthCheckTimeout: "",
						Fips:                           nil,
						HostClusterConfig: &models.V1HostClusterConfig{
							ClusterEndpoint: &models.V1HostClusterEndpoint{
								Config: nil,
								Type:   "LoadBalancer",
							},
							ClusterGroup:  nil,
							HostCluster:   nil,
							IsHostCluster: isHost,
						},
						LifecycleConfig:             nil,
						MachineHealthConfig:         nil,
						MachineManagementConfig:     nil,
						UpdateWorkerPoolsInParallel: false,
					},
					ClusterProfileTemplates: nil,
					ClusterType:             "",
				},
				Status: &models.V1SpectroClusterStatus{
					State: "running",
				},
			}
			return cluster, nil
		},
	}

	ctx := context.Background()
	resourceKubevirtDataVolumeCreate(ctx, rd, m)
}

func TestDeleteDataVolume(t *testing.T) {
	assert := assert.New(t)
	rd := prepareDataVolumeTestData()

	m := &client.V1Client{
		DeleteDataVolumeFn: func(uid string, namespace string, name string, body *models.V1VMRemoveVolumeEntity) error {
			if uid != "cluster-123" {
				return errors.New("unexpected cluster_uid")
			}
			if namespace != "default" {
				return errors.New("unexpected vm_namespace")
			}
			if name != "vm-test" {
				return errors.New("unexpected vm_name")
			}
			if *body.RemoveVolumeOptions.Name != "vol-test" {
				return errors.New("unexpected volume name")
			}
			return nil
		},
	}

	ctx := context.Background()
	diags := resourceKubevirtDataVolumeDelete(ctx, rd, m)
	if diags.HasError() {
		assert.Error(errors.New("delete operation failed"))
	} else {
		assert.NoError(nil)
	}
}

func TestReadDataVolumeWithoutStatus(t *testing.T) {
	assert := assert.New(t)
	rd := prepareDataVolumeTestData()
	rd.SetId("cluster-123/default/vm-test/vol-test")
	m := &client.V1Client{
		GetVirtualMachineWithoutStatusFn: func(uid string) (*models.V1ClusterVirtualMachine, error) {
			if uid != "cluster-123" {
				return nil, errors.New("unexpected cluster_uid")
			}

			// Note: we added another data volume template here to cover the for loop in the resourceKubevirtDataVolumeRead function
			return &models.V1ClusterVirtualMachine{
				Spec: &models.V1ClusterVirtualMachineSpec{
					DataVolumeTemplates: []*models.V1VMDataVolumeTemplateSpec{
						{
							Metadata: &models.V1VMObjectMeta{
								Name:      "vol-test",
								Namespace: "default",
							},
							Spec: &models.V1VMDataVolumeSpec{
								Checkpoints:       []*models.V1VMDataVolumeCheckpoint{}, // Fill this with appropriate values if required
								ContentType:       "kubevirt",
								FinalCheckpoint:   true,
								Preallocation:     true,
								PriorityClassName: "high-priority",
								Pvc:               &models.V1VMPersistentVolumeClaimSpec{
									// Fill this with appropriate values
								},
								Source: &models.V1VMDataVolumeSource{
									// Fill this with appropriate values
								},
								SourceRef: &models.V1VMDataVolumeSourceRef{
									// Fill this with appropriate values
								},
								Storage: &models.V1VMStorageSpec{
									// Fill this with appropriate values
								},
							},
						},
					},
				},
			}, nil
		},
	}

	ctx := context.Background()
	diags := resourceKubevirtDataVolumeRead(ctx, rd, m)
	if diags.HasError() {
		assert.Error(errors.New("read operation failed"))
	} else {
		assert.NoError(nil)
	}

	// Read from metadata block
	metadata := rd.Get("metadata").([]interface{})[0].(map[string]interface{})

	// Check that the resource data has been updated correctly
	assert.Equal("vol-test", metadata["name"])
	assert.Equal("default", metadata["namespace"])
}

func TestReadDataVolume(t *testing.T) {
	assert := assert.New(t)
	rd := prepareDataVolumeTestData()

	m := &client.V1Client{
		GetVirtualMachineFn: func(uid string) (*models.V1ClusterVirtualMachine, error) {
			if uid != "cluster-123" {
				return nil, errors.New("unexpected cluster_uid")
			}

			return &models.V1ClusterVirtualMachine{
				Spec: &models.V1ClusterVirtualMachineSpec{
					DataVolumeTemplates: []*models.V1VMDataVolumeTemplateSpec{
						{
							Metadata: &models.V1VMObjectMeta{
								Name:      "vol-test",
								Namespace: "default",
							},
						},
					},
				},
			}, nil
		},
	}

	ctx := context.Background()
	diags := resourceKubevirtDataVolumeRead(ctx, rd, m)
	if diags.HasError() {
		assert.Error(errors.New("read operation failed"))
	} else {
		assert.NoError(nil)
	}
}

func TestExpandAddVolumeOptions(t *testing.T) {
	assert := assert.New(t)

	addVolumeOptions := []interface{}{
		map[string]interface{}{
			"name": "test-volume",
			"disk": []interface{}{
				map[string]interface{}{
					"name": "test-disk",
					"bus":  "scsi",
				},
			},
			"volume_source": []interface{}{
				map[string]interface{}{
					"data_volume": []interface{}{
						map[string]interface{}{
							"name":         "test-data-volume",
							"hotpluggable": true,
						},
					},
				},
			},
		},
	}

	expected := &models.V1VMAddVolumeOptions{
		Name: types.Ptr("test-volume"),
		Disk: &models.V1VMDisk{
			Name: types.Ptr("test-disk"),
			Disk: &models.V1VMDiskTarget{
				Bus: "scsi",
			},
		},
		VolumeSource: &models.V1VMHotplugVolumeSource{
			DataVolume: &models.V1VMCoreDataVolumeSource{
				Name:         types.Ptr("test-data-volume"),
				Hotpluggable: true,
			},
		},
	}

	result := ExpandAddVolumeOptions(addVolumeOptions)

	assert.Equal(expected, result, "ExpandAddVolumeOptions returned unexpected result")
}
