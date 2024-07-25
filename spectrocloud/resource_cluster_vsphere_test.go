package spectrocloud

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func prepareClusterVsphereTestData() *schema.ResourceData {
	d := resourceClusterVsphere().TestResourceData()

	d.SetId("")
	d.Set("name", "vsphere-picard-2")
	cConfig := make([]map[string]interface{}, 0)
	cConfig = append(cConfig, map[string]interface{}{
		"id": "vmware-basic-infra-profile-id",
	})
	d.Set("cluster_meta_attribute", "{'nic_name': 'test', 'env': 'stage'}")
	d.Set("cluster_profile", cConfig)
	d.Set("cloud_account_id", "vmware-basic-account-id")

	// Cluster Rbac binding
	// rbacRole := make([]interface{}, 0)
	// rbacRole = append(rbacRole, map[string]interface{}{
	// 	"kind": "Role",
	// 	"name": "testUserRoleFromNS",
	// })
	// rbacSubjects := make([]interface{}, 0)
	// rbacSubjects = append(rbacSubjects, map[string]interface{}{
	// 	"type":      "Role",
	// 	"name":      "testUserRoleFromNS",
	// 	"namespace": "testrolenamespace",
	// })
	// rbac := make([]map[string]interface{}, 0)
	// r := map[string]interface{}{
	// 	"type":      "RoleBinding1",
	// 	"namespace": "test5ns",
	// 	"role":      rbacRole,
	// 	"subjects":  rbacSubjects,
	// }
	// rbac = append(rbac, r)

	// cloud config
	keys := []string{"SSHKey1", "SSHKey2"}
	cloudConfig := make([]map[string]interface{}, 0)
	con := map[string]interface{}{
		"ssh_keys":              keys,
		"datacenter":            "Datacenter",
		"folder":                "sc_test/terraform",
		"network_type":          "DDNS",
		"network_search_domain": "spectrocloud.dev",
	}
	cloudConfig = append(cloudConfig, con)
	d.Set("cloud_config", cloudConfig)

	mPools := make([]map[string]interface{}, 0)

	// Adding control-plane pool
	cpPlacement := make([]interface{}, 0)
	cpPlacement = append(cpPlacement, map[string]interface{}{
		"id":                "",
		"cluster":           "test cluster",
		"resource_pool":     "Default",
		"datastore":         "datastore55_2",
		"network":           "VM Network",
		"static_ip_pool_id": "testpoolid",
	})
	cpInstance := make([]interface{}, 0)
	cpInstance = append(cpInstance, map[string]interface{}{
		"disk_size_gb": 40,
		"memory_mb":    8192,
		"cpu":          4,
	})
	mPools = append(mPools, map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"name":                    "cp-pool",
		"count":                   1,
		"placement":               cpPlacement,
		"instance_type":           cpInstance,
		"node":                    []interface{}{},
	})

	// Adding Worker pool
	workerPlacement := make([]interface{}, 0)
	workerPlacement = append(workerPlacement, map[string]interface{}{
		"id":                "",
		"cluster":           "test cluster",
		"resource_pool":     "Default",
		"datastore":         "datastore55_2",
		"network":           "VM Network",
		"static_ip_pool_id": "testpoolid",
	})

	workerInstance := make([]interface{}, 0)
	workerInstance = append(workerInstance, map[string]interface{}{
		"disk_size_gb": 40,
		"memory_mb":    8192,
		"cpu":          4,
	})

	mPools = append(mPools, map[string]interface{}{
		"control_plane":           false,
		"control_plane_as_worker": false,
		"name":                    "worker-basic",
		"count":                   1,
		"min":                     1,
		"max":                     3,
		"placement":               workerPlacement,
		"instance_type":           workerInstance,
		"node":                    []interface{}{},
	})
	d.Set("machine_pool", mPools)
	return d
}

func TestToVsphereCluster(t *testing.T) {
	assert := assert.New(t)
	// Create a mock ResourceData object
	d := prepareClusterVsphereTestData()

	m := &client.V1Client{
		GetClusterWithoutStatusFn: func(uid string) (*models.V1SpectroCluster, error) {
			return &models.V1SpectroCluster{
				Metadata: nil,
				Spec:     nil,
			}, nil
		},
	}

	vSphereSchema, err := toVsphereCluster(m, d)
	assert.Nil(err)

	// Check the output against the expected values
	// Verifying cluster name attribute
	assert.Equal(d.Get("name"), vSphereSchema.Metadata.Name)

	// Verifying cluster name attribute
	assert.Equal("vmware-basic-infra-profile-id", vSphereSchema.Spec.Profiles[0].UID)

	// Verifying cluster_meta_attribute attribute
	assert.Equal("{'nic_name': 'test', 'env': 'stage'}", vSphereSchema.Spec.ClusterConfig.ClusterMetaAttribute)

	// Verifying account id attribute
	assert.Equal("vmware-basic-account-id", vSphereSchema.Spec.CloudAccountUID)

	// Verifying cloud config attributes
	assert.Equal("spectrocloud.dev", vSphereSchema.Spec.CloudConfig.ControlPlaneEndpoint.DdnsSearchDomain)
	assert.Equal("DDNS", vSphereSchema.Spec.CloudConfig.ControlPlaneEndpoint.Type)
	assert.Equal("Datacenter", vSphereSchema.Spec.CloudConfig.Placement.Datacenter)
	assert.Equal("sc_test/terraform", vSphereSchema.Spec.CloudConfig.Placement.Folder)
	assert.Equal(2, len(vSphereSchema.Spec.CloudConfig.SSHKeys))
	assert.Equal(false, vSphereSchema.Spec.CloudConfig.StaticIP)

	// Verifying control-plane pool attributes
	assert.Equal(2, len(vSphereSchema.Spec.Machinepoolconfig))
	cpPoolIndex := 0
	workerPoolIndex := 1
	if *vSphereSchema.Spec.Machinepoolconfig[0].PoolConfig.Name == "cp-pool" {
		cpPoolIndex = 0
		workerPoolIndex = 1
	} else {
		cpPoolIndex = 1
		workerPoolIndex = 0
	}

	assert.Equal("cp-pool", *vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].PoolConfig.Name)
	assert.Equal(true, vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].PoolConfig.IsControlPlane)
	assert.Equal(int32(40), *vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.InstanceType.DiskGiB)
	assert.Equal(int64(8192), *vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.InstanceType.MemoryMiB)
	assert.Equal(int32(4), *vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.InstanceType.NumCPUs)
	assert.Equal("test cluster", vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.Placements[0].Cluster)
	assert.Equal("datastore55_2", vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.Placements[0].Datastore)
	assert.Equal("Default", vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.Placements[0].ResourcePool)
	assert.Equal("VM Network", *vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.Placements[0].Network.NetworkName)
	assert.Equal("testpoolid", vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.Placements[0].Network.ParentPoolUID)
	assert.Equal(true, vSphereSchema.Spec.Machinepoolconfig[cpPoolIndex].CloudConfig.Placements[0].Network.StaticIP)

	// Verifying Worker pool attributes
	assert.Equal("worker-basic", *vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].PoolConfig.Name)
	assert.Equal(false, vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].PoolConfig.IsControlPlane)
	assert.Equal(int32(40), *vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.InstanceType.DiskGiB)
	assert.Equal(int64(8192), *vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.InstanceType.MemoryMiB)
	assert.Equal(int32(4), *vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.InstanceType.NumCPUs)
	assert.Equal("test cluster", vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.Placements[0].Cluster)
	assert.Equal("datastore55_2", vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.Placements[0].Datastore)
	assert.Equal("Default", vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.Placements[0].ResourcePool)
	assert.Equal("VM Network", *vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.Placements[0].Network.NetworkName)
	assert.Equal("testpoolid", vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.Placements[0].Network.ParentPoolUID)
	assert.Equal(true, vSphereSchema.Spec.Machinepoolconfig[workerPoolIndex].CloudConfig.Placements[0].Network.StaticIP)
}

func TestToCloudConfigUpdate(t *testing.T) {
	assert := assert.New(t)
	cloudConfig := map[string]interface{}{
		"ssh_key":               "ssh-rsa AAAAB3NzaC1y",
		"datacenter":            "Datacenter",
		"folder":                "sc_test/terraform",
		"network_type":          "DDNS",
		"network_search_domain": "spectrocloud.dev",
		"static_ip":             false,
	}
	cloudEntity := toCloudConfigUpdate(cloudConfig)
	assert.Equal("spectrocloud.dev", cloudEntity.ClusterConfig.ControlPlaneEndpoint.DdnsSearchDomain)
	assert.Equal("DDNS", cloudEntity.ClusterConfig.ControlPlaneEndpoint.Type)
	assert.Equal("Datacenter", cloudEntity.ClusterConfig.Placement.Datacenter)
	assert.Equal("sc_test/terraform", cloudEntity.ClusterConfig.Placement.Folder)
	assert.Equal("spectro-templates", cloudEntity.ClusterConfig.Placement.ImageTemplateFolder)
	assert.Equal("ssh-rsa AAAAB3NzaC1y", cloudEntity.ClusterConfig.SSHKeys[0])
	assert.Equal(false, cloudEntity.ClusterConfig.StaticIP)
}

func TestResourceClusterVsphereCreate(t *testing.T) {

	// Create a mock ResourceData object
	d := prepareClusterVsphereTestData()
	d.Set("skip_completion", true)
	m := &client.V1Client{
		CreateClusterVsphereFn: func(cluster *models.V1SpectroVsphereClusterEntity) (string, error) {
			return "vsphere-cluster-uid", nil
		},
	}
	ctx := context.Background()
	diags := resourceClusterVsphereCreate(ctx, d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}

	if d.Id() != "vsphere-cluster-uid" {
		t.Errorf("Expected ID to be 'test-uid', got %s", d.Id())
	}
}

func TestResourceClusterVsphereCreateError(t *testing.T) {

	d := prepareClusterVsphereTestData()
	d.Set("skip_completion", true)
	m := &client.V1Client{
		CreateClusterVsphereFn: func(cluster *models.V1SpectroVsphereClusterEntity) (string, error) {
			return "", errors.New("covering error case")
		},
	}
	ctx := context.Background()
	diags := resourceClusterVsphereCreate(ctx, d, m)
	if diags[0].Summary != "covering error case" {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func getClientForCluster() *client.V1Client {
	m := &client.V1Client{
		GetCloudConfigVsphereFn: func(uid string) (*models.V1VsphereCloudConfig, error) {
			return getCloudConfig(), nil
		},
		GetClusterFn: func(scope, uid string) (*models.V1SpectroCluster, error) {
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
					CloudType: "vsphere",
					ClusterConfig: &models.V1ClusterConfig{
						ClusterRbac:                    nil,
						ClusterResources:               nil,
						ControlPlaneHealthCheckTimeout: "",
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
					Repave: &models.V1ClusterRepaveStatus{
						State: "",
					},
				},
			}
			return cluster, nil
		},
		GetClusterBackupConfigFn: func(uid string) (*models.V1ClusterBackup, error) {
			clusterBackup := &models.V1ClusterBackup{
				Metadata: nil,
				Spec: &models.V1ClusterBackupSpec{
					ClusterUID: "vsphere-cluster-uid",
					Config: &models.V1ClusterBackupConfig{
						BackupLocationUID:       "test-back-uid",
						BackupName:              "unit-back",
						BackupPrefix:            "vsphere",
						DurationInHours:         3,
						IncludeAllDisks:         false,
						IncludeClusterResources: false,
						LocationType:            "",
						Namespaces:              nil,
						Schedule: &models.V1ClusterFeatureSchedule{
							ScheduledRunTime: "daily",
						},
					},
				},
				Status: nil,
			}
			return clusterBackup, nil
		},
		GetClusterKubeConfigFn: func(uid string) (string, error) {
			return "testKubeConfig", nil
		},
		GetClusterAdminConfigFn: func(uid string) (string, error) {
			return "testAdminKubeConfig", nil
		},
		GetClusterScanConfigFn: func(uid string) (*models.V1ClusterComplianceScan, error) {
			clusterCom := &models.V1ClusterComplianceScan{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "vsphere-cluster",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "conpli-uid",
				},
				Spec: &models.V1ClusterComplianceScanSpec{
					ClusterUID: "vsphere-cluster-uid",
					DriverSpec: map[string]models.V1ComplianceScanDriverSpec{
						"kube-bench": {
							Config: &models.V1ComplianceScanConfig{
								Schedule: &models.V1ClusterFeatureSchedule{
									ScheduledRunTime: "daily",
								},
							},
							IsClusterConfig: false,
						},
						"kube-hunter": {
							Config: &models.V1ComplianceScanConfig{
								Schedule: &models.V1ClusterFeatureSchedule{
									ScheduledRunTime: "daily",
								},
							},
							IsClusterConfig: false,
						},
						"sonobuoy": {
							Config: &models.V1ComplianceScanConfig{
								Schedule: &models.V1ClusterFeatureSchedule{
									ScheduledRunTime: "daily",
								},
							},
							IsClusterConfig: false,
						},
					},
				},
			}
			return clusterCom, nil
		},
		GetClusterRbacConfigFn: func(uid string) (*models.V1ClusterRbacs, error) {
			var rbacs []*models.V1ClusterRbac
			var subject []*models.V1ClusterRbacSubjects
			var bindings []*models.V1ClusterRbacBinding
			subject = append(subject, &models.V1ClusterRbacSubjects{
				Name:      "test-subject",
				Namespace: "vsphere-test",
				Type:      "test-subject",
			})
			bindings = append(bindings, &models.V1ClusterRbacBinding{
				Namespace: "vsphere-unittest",
				Role: &models.V1ClusterRoleRef{
					Kind: "scan",
					Name: "test-kind",
				},
				Subjects: subject,
				Type:     "test",
			})
			rbacs = append(rbacs, &models.V1ClusterRbac{
				Metadata: nil,
				Spec: &models.V1ClusterRbacSpec{
					Bindings:      bindings,
					RelatedObject: nil,
				},
				Status: nil,
			})
			clusterRbac := &models.V1ClusterRbacs{
				Items: rbacs,
			}
			return clusterRbac, nil
		},
		GetClusterNamespaceConfigFn: func(uid string) (*models.V1ClusterNamespaceResources, error) {
			var nResources []*models.V1ClusterNamespaceResource
			nResources = append(nResources, &models.V1ClusterNamespaceResource{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-namespace-unit",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "",
				},
				Spec: &models.V1ClusterNamespaceSpec{
					IsRegex:       false,
					RelatedObject: nil,
					ResourceAllocation: &models.V1ClusterNamespaceResourceAllocation{
						CPUCores:  5,
						MemoryMiB: 1234,
					},
				},
				Status: nil,
			})
			namespaceResource := &models.V1ClusterNamespaceResources{
				Items: nResources,
			}
			return namespaceResource, nil
		},
		GetClusterWithoutStatusFn: func(uid string) (*models.V1SpectroCluster, error) {
			cluster := &models.V1SpectroCluster{
				Metadata: nil,
			}
			cluster.Status = &models.V1SpectroClusterStatus{
				AbortTimestamp: models.V1Time{},
				AddOnServices:  nil,
				APIEndpoints:   nil,
				ClusterImport:  nil,
				Conditions:     nil,
				Location: &models.V1ClusterLocation{
					CountryCode: "IN",
					CountryName: "India",
					GeoLoc: &models.V1GeolocationLatlong{
						Latitude:  12.4241231,
						Longitude: 1932.12312,
					},
					RegionCode: "12",
					RegionName: "Asia",
				},
				Packs:         nil,
				ProfileStatus: nil,
				Services:      nil,
				SpcApply:      nil,
				State:         "running",
				Upgrades:      nil,
				Virtual:       nil,
			}
			return cluster, nil
		},
	}
	return m
}
func TestResourceClusterVsphereRead(t *testing.T) {
	// Create a mock ResourceData object
	d := prepareClusterVsphereTestData()
	m := getClientForCluster()
	ctx := context.Background()
	diags := resourceClusterVsphereRead(ctx, d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceClusterVsphereReadValidationErrorSpec(t *testing.T) {
	d := prepareClusterVsphereTestData()
	m := &client.V1Client{
		GetClusterWithoutStatusFn: func(uid string) (*models.V1SpectroCluster, error) {
			cluster := &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					UID: "mockid123",
				},
				Spec: nil,
			}
			cluster.Status = &models.V1SpectroClusterStatus{
				State: "running",
			}
			return cluster, nil
		},
	}
	ctx := context.Background()
	diags := resourceClusterVsphereRead(ctx, d, m)
	if len(diags) == 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	if diags[0].Summary != "cluster spec is nil in cluster mockid123" {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceClusterVsphereReadValidationErrorCloudType(t *testing.T) {
	d := prepareClusterVsphereTestData()
	m := &client.V1Client{
		GetClusterWithoutStatusFn: func(uid string) (*models.V1SpectroCluster, error) {
			cluster := &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					UID: "mockid123",
				},
				Spec: &models.V1SpectroClusterSpec{
					CloudType: "aws", // wrong cloud type, vsphere expected
				},
			}
			cluster.Status = &models.V1SpectroClusterStatus{
				State: "running",
			}
			return cluster, nil
		},
	}
	ctx := context.Background()
	diags := resourceClusterVsphereRead(ctx, d, m)
	if len(diags) == 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	if diags[0].Summary != "resource with id mockid123 is not of type spectrocloud_cluster_vsphere, need to correct resource type" {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func TestResourceClusterVsphereReadNilCluster(t *testing.T) {
	// Create a mock ResourceData object
	d := prepareClusterVsphereTestData()
	m := &client.V1Client{
		GetClusterFn: func(scope, uid string) (*models.V1SpectroCluster, error) {
			return nil, nil
		},
	}
	ctx := context.Background()
	diags := resourceClusterVsphereRead(ctx, d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	assert.Equal(t, "", d.Id())
}

func TestResourceClusterVsphereReadError(t *testing.T) {
	// Create a mock ResourceData object
	d := prepareClusterVsphereTestData()
	m := &client.V1Client{
		GetClusterFn: func(scope, uid string) (*models.V1SpectroCluster, error) {
			return nil, errors.New("unexpected Error")
		},
	}
	ctx := context.Background()
	diags := resourceClusterVsphereRead(ctx, d, m)
	assert.Equal(t, len(diags), 1)
	if diags[0].Summary != "unexpected Error" {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
}

func getMachinePlacement() []*models.V1VspherePlacementConfig {
	network := new(string)
	*network = "test-net"
	var placement []*models.V1VspherePlacementConfig
	placement = append(placement, &models.V1VspherePlacementConfig{
		Cluster:             "test-cluster",
		Datacenter:          "vsphere",
		Datastore:           "vcenter",
		Folder:              "/test/",
		ImageTemplateFolder: "",
		Network: &models.V1VsphereNetworkConfig{
			IPPool:      nil,
			NetworkName: network,
			ParentPoolRef: &models.V1ObjectReference{
				UID: "test-pool-id",
			},
			StaticIP: false,
		},
		ResourcePool:      "",
		StoragePolicyName: "",
		UID:               "test-uid",
	})
	return placement
}

func getMPools() []*models.V1VsphereMachinePoolConfig {
	var mTaint []*models.V1Taint
	diskGb := new(int32)
	*diskGb = 23
	memMb := new(int64)
	*memMb = 120
	numCpu := new(int32)
	*numCpu = 2
	mTaint = append(mTaint, &models.V1Taint{
		Effect:    "start",
		Key:       "owner",
		TimeAdded: models.V1Time{},
		Value:     "siva",
	})
	var mPool []*models.V1VsphereMachinePoolConfig
	mPool = append(mPool, &models.V1VsphereMachinePoolConfig{
		AdditionalLabels: map[string]string{
			"type":  "unittest",
			"owner": "siva",
		},
		AdditionalTags: nil,
		InstanceType: &models.V1VsphereInstanceType{
			DiskGiB:   diskGb,
			MemoryMiB: memMb,
			NumCPUs:   numCpu,
		},
		IsControlPlane: nil,
		Labels:         nil,
		MaxSize:        0,
		MinSize:        0,
		Name:           "",
		Placements:     getMachinePlacement(),
		Size:           0,
		Taints:         mTaint,
		UpdateStrategy: &models.V1UpdateStrategy{
			Type: "",
		},
		UseControlPlaneAsWorker: false,
	})
	return mPool
}

func getCloudConfig() *models.V1VsphereCloudConfig {
	cloudConfig := &models.V1VsphereCloudConfig{
		APIVersion: "v1",
		Kind:       "",
		Metadata:   nil,
		Spec: &models.V1VsphereCloudConfigSpec{
			CloudAccountRef: &models.V1ObjectReference{
				APIVersion:      "",
				FieldPath:       "",
				Kind:            "",
				Name:            "",
				Namespace:       "",
				ResourceVersion: "",
				UID:             "vmware-basic-account-id",
			},
			ClusterConfig:     nil,
			EdgeHostRef:       nil,
			MachinePoolConfig: getMPools(),
		},
		Status: nil,
	}
	return cloudConfig
}

func TestFlattenCloudConfigVsphere(t *testing.T) {
	d := prepareClusterVsphereTestData()
	m := &client.V1Client{
		GetCloudConfigVsphereFn: func(uid string) (*models.V1VsphereCloudConfig, error) {
			return getCloudConfig(), nil
		},
		GetCloudConfigVsphereValuesFn: func(uid string) (*models.V1VsphereCloudConfig, error) {
			vsphereConfig := &models.V1VsphereCloudConfig{
				APIVersion: "v1",
				Kind:       "",
				Metadata:   nil,
				Spec: &models.V1VsphereCloudConfigSpec{
					CloudAccountRef: nil,
					ClusterConfig: &models.V1VsphereClusterConfig{
						ControlPlaneEndpoint: &models.V1ControlPlaneEndPoint{
							DdnsSearchDomain: "spectro.dev.com",
							Host:             "spectro.dev",
							Type:             "test",
						},
						NtpServers: strings.Split("pool.ntp.org],", ","),
						Placement: &models.V1VspherePlacementConfig{
							Cluster:             "",
							Datacenter:          "vpshere",
							Datastore:           "testing",
							Folder:              "test/unit/test",
							ImageTemplateFolder: "",
							Network:             nil,
							ResourcePool:        "",
							StoragePolicyName:   "",
							UID:                 "",
						},
						SSHKeys:  []string{"ssh -A asdfasdf"},
						StaticIP: false,
					},
					EdgeHostRef:       nil,
					MachinePoolConfig: getMPools(),
				},
			}
			return vsphereConfig, nil

		},
	}
	diags := flattenCloudConfigVsphere("", d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}

}

func TestFlattenClusterConfigsVsphere(t *testing.T) {
	inputCloudConfig := &models.V1VsphereCloudConfig{
		Spec: &models.V1VsphereCloudConfigSpec{
			ClusterConfig: &models.V1VsphereClusterConfig{
				SSHKeys:    []string{"SSHKey1", "SSHKey1"},
				StaticIP:   true,
				NtpServers: []string{"ntp1", "ntp2"},
				Placement: &models.V1VspherePlacementConfig{
					Datacenter: "Datacenter1",
					Folder:     "Folder1",
				},
				ControlPlaneEndpoint: &models.V1ControlPlaneEndPoint{
					Type:             "VIP",
					DdnsSearchDomain: "example.dev",
				},
			},
		},
	}
	d := prepareClusterVsphereTestData()
	flattenedConfig := flattenClusterConfigsVsphere(d, inputCloudConfig)

	flattenedConfigMap := flattenedConfig.([]interface{})[0].(map[string]interface{})
	if flattenedConfigMap["datacenter"].(string) != inputCloudConfig.Spec.ClusterConfig.Placement.Datacenter {
		t.Errorf("Failed to flatten 'datacenter' field correctly")
	}
	if flattenedConfigMap["folder"].(string) != inputCloudConfig.Spec.ClusterConfig.Placement.Folder {
		t.Errorf("Failed to flatten 'folder' field correctly")
	}
	if !reflect.DeepEqual(flattenedConfigMap["ssh_keys"].([]string), inputCloudConfig.Spec.ClusterConfig.SSHKeys) {
		t.Errorf("Failed to flatten 'ssh_keys' field correctly")
	}
	if flattenedConfigMap["static_ip"].(bool) != inputCloudConfig.Spec.ClusterConfig.StaticIP {
		t.Errorf("Failed to flatten 'static_ip' field correctly")
	}
	if flattenedConfigMap["network_type"].(string) != inputCloudConfig.Spec.ClusterConfig.ControlPlaneEndpoint.Type {
		t.Errorf("Failed to flatten 'network_type' field correctly")
	}
	if flattenedConfigMap["network_search_domain"].(string) != inputCloudConfig.Spec.ClusterConfig.ControlPlaneEndpoint.DdnsSearchDomain {
		t.Errorf("Failed to flatten 'network_search_domain' field correctly")
	}
	flattenedNtpServers := flattenedConfigMap["ntp_servers"].([]string)
	if !reflect.DeepEqual(flattenedNtpServers, inputCloudConfig.Spec.ClusterConfig.NtpServers) {
		t.Errorf("Failed to flatten 'ntp_servers' field correctly")
	}
}

func TestFlattenClusterConfigsVsphereNil(t *testing.T) {
	d := prepareClusterVsphereTestData()
	flatCloudConfig := flattenClusterConfigsVsphere(d, nil)
	if flatCloudConfig == nil {
		t.Errorf("flattenClusterConfigsVsphere returning value for nill: %#v", flatCloudConfig)
	}
}

func TestFlattenMachinePoolConfigsVsphereNil(t *testing.T) {
	flatPool := flattenMachinePoolConfigsVsphere(nil)
	if len(flatPool) > 0 {
		t.Errorf("flattenMachinePoolConfigsVsphere returning value for nill: %#v", flatPool)
	}
}

func TestResourceClusterVsphereUpdate(t *testing.T) {
	d := prepareClusterVsphereTestData()
	m := &client.V1Client{
		GetClusterFn: func(scope, uid string) (*models.V1SpectroCluster, error) {
			return nil, nil
		},
		GetCloudConfigVsphereFn: func(cloudConfigUid string) (*models.V1VsphereCloudConfig, error) {
			return nil, nil
		},
	}
	ctx := context.Background()
	diags := resourceClusterVsphereUpdate(ctx, d, m)
	if len(diags) > 0 {
		t.Errorf("Unexpected diagnostics: %#v", diags)
	}
	assert.Equal(t, "", d.Id())

}
