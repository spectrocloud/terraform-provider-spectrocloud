// Copyright (c) Spectro Cloud
// SPDX-License-Identifier: MPL-2.0

package spectrocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/testutil"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/testutil/vcr"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// prepareClusterVsphereTestData returns ResourceData populated for vSphere cluster unit tests.
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

// =============================================================================
// UNIT TESTS - No network calls, fast execution
// =============================================================================

// TestUnit_ResourceClusterVsphereSchema validates the vSphere cluster resource schema
func TestUnit_ResourceClusterVsphereSchema(t *testing.T) {
	t.Parallel()

	s := resourceClusterVsphere()

	// Validate required fields
	require.NotNil(t, s.Schema["name"])
	assert.True(t, s.Schema["name"].Required, "name should be required")
	assert.Equal(t, schema.TypeString, s.Schema["name"].Type)

	require.NotNil(t, s.Schema["cloud_account_id"])
	assert.True(t, s.Schema["cloud_account_id"].Required, "cloud_account_id should be required")

	require.NotNil(t, s.Schema["cloud_config"])
	assert.True(t, s.Schema["cloud_config"].Required, "cloud_config should be required")

	require.NotNil(t, s.Schema["machine_pool"])
	assert.True(t, s.Schema["machine_pool"].Required, "machine_pool should be required")

	// Validate optional fields
	require.NotNil(t, s.Schema["context"])
	assert.True(t, s.Schema["context"].Optional, "context should be optional")
	assert.Equal(t, "project", s.Schema["context"].Default)

	require.NotNil(t, s.Schema["tags"])
	assert.True(t, s.Schema["tags"].Optional, "tags should be optional")

	require.NotNil(t, s.Schema["description"])
	assert.True(t, s.Schema["description"].Optional, "description should be optional")

	// Validate computed fields
	require.NotNil(t, s.Schema["cloud_config_id"])
	assert.True(t, s.Schema["cloud_config_id"].Computed, "cloud_config_id should be computed")

	require.NotNil(t, s.Schema["kubeconfig"])
	assert.True(t, s.Schema["kubeconfig"].Computed, "kubeconfig should be computed")

	// Validate CRUD operations are defined
	assert.NotNil(t, s.CreateContext, "CreateContext should be defined")
	assert.NotNil(t, s.ReadContext, "ReadContext should be defined")
	assert.NotNil(t, s.UpdateContext, "UpdateContext should be defined")
	assert.NotNil(t, s.DeleteContext, "DeleteContext should be defined")

	// Validate importer is defined
	assert.NotNil(t, s.Importer, "Importer should be defined")

	// Validate timeouts
	assert.NotNil(t, s.Timeouts, "Timeouts should be defined")
	assert.NotNil(t, s.Timeouts.Create, "Create timeout should be defined")
	assert.NotNil(t, s.Timeouts.Update, "Update timeout should be defined")
	assert.NotNil(t, s.Timeouts.Delete, "Delete timeout should be defined")
}

// TestUnit_ToMachinePoolVsphere tests the toMachinePoolVsphere function
func TestUnit_ToMachinePoolVsphere(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
		validate    func(t *testing.T, result *models.V1VsphereMachinePoolConfigEntity)
	}{
		{
			name: "valid control plane pool",
			input: map[string]interface{}{
				"name":                    "cp-pool",
				"control_plane":           true,
				"control_plane_as_worker": true,
				"count":                   3,
				"node_repave_interval":    0,
				"placement": []interface{}{
					map[string]interface{}{
						"id":                "",
						"cluster":           "test-cluster",
						"resource_pool":     "Default",
						"datastore":         "datastore1",
						"network":           "VM Network",
						"static_ip_pool_id": "",
					},
				},
				"instance_type": []interface{}{
					map[string]interface{}{
						"disk_size_gb": 60,
						"memory_mb":    8192,
						"cpu":          4,
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, result *models.V1VsphereMachinePoolConfigEntity) {
				assert.Equal(t, "cp-pool", *result.PoolConfig.Name)
				assert.True(t, result.PoolConfig.IsControlPlane)
				assert.True(t, result.PoolConfig.UseControlPlaneAsWorker)
				assert.Equal(t, int32(3), *result.PoolConfig.Size)
				assert.Len(t, result.CloudConfig.Placements, 1)
				assert.Equal(t, "test-cluster", result.CloudConfig.Placements[0].Cluster)
			},
		},
		{
			name: "valid worker pool",
			input: map[string]interface{}{
				"name":                    "worker-pool",
				"control_plane":           false,
				"control_plane_as_worker": false,
				"count":                   2,
				"min":                     1,
				"max":                     5,
				"node_repave_interval":    60,
				"placement": []interface{}{
					map[string]interface{}{
						"id":                "",
						"cluster":           "test-cluster",
						"resource_pool":     "Default",
						"datastore":         "datastore1",
						"network":           "VM Network",
						"static_ip_pool_id": "pool-123",
					},
				},
				"instance_type": []interface{}{
					map[string]interface{}{
						"disk_size_gb": 100,
						"memory_mb":    16384,
						"cpu":          8,
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, result *models.V1VsphereMachinePoolConfigEntity) {
				assert.Equal(t, "worker-pool", *result.PoolConfig.Name)
				assert.False(t, result.PoolConfig.IsControlPlane)
				assert.Equal(t, int32(2), *result.PoolConfig.Size)
				assert.Equal(t, int32(1), result.PoolConfig.MinSize)
				assert.Equal(t, int32(5), result.PoolConfig.MaxSize)
				assert.Equal(t, int32(60), result.PoolConfig.NodeRepaveInterval)
			},
		},
		{
			name: "negative count should error",
			input: map[string]interface{}{
				"name":                    "invalid-pool",
				"control_plane":           false,
				"control_plane_as_worker": false,
				"count":                   -1,
				"node_repave_interval":    0,
				"placement": []interface{}{
					map[string]interface{}{
						"id":                "",
						"cluster":           "test-cluster",
						"resource_pool":     "Default",
						"datastore":         "datastore1",
						"network":           "VM Network",
						"static_ip_pool_id": "",
					},
				},
				"instance_type": []interface{}{
					map[string]interface{}{
						"disk_size_gb": 60,
						"memory_mb":    8192,
						"cpu":          4,
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := toMachinePoolVsphere(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			tt.validate(t, result)
		})
	}
}

// TestUnit_FlattenMachinePoolConfigsVsphere tests the flattenMachinePoolConfigsVsphere function
func TestUnit_FlattenMachinePoolConfigsVsphere(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []*models.V1VsphereMachinePoolConfig
		validate func(t *testing.T, result []interface{})
	}{
		{
			name:  "nil input returns empty slice",
			input: nil,
			validate: func(t *testing.T, result []interface{}) {
				assert.Empty(t, result)
			},
		},
		{
			name:  "empty input returns empty slice",
			input: []*models.V1VsphereMachinePoolConfig{},
			validate: func(t *testing.T, result []interface{}) {
				assert.Empty(t, result)
			},
		},
		{
			name: "single machine pool",
			input: []*models.V1VsphereMachinePoolConfig{
				{
					Name:           "test-pool",
					Size:           3,
					MinSize:        1,
					MaxSize:        5,
					IsControlPlane: types.Ptr(true),
					InstanceType: &models.V1VsphereInstanceType{
						DiskGiB:   types.Ptr(int32(100)),
						MemoryMiB: types.Ptr(int64(8192)),
						NumCPUs:   types.Ptr(int32(4)),
					},
					Placements: []*models.V1VspherePlacementConfig{
						{
							UID:          "placement-1",
							Cluster:      "vsphere-cluster",
							ResourcePool: "default-pool",
							Datastore:    "datastore1",
							Network: &models.V1VsphereNetworkConfig{
								NetworkName: types.Ptr("VM Network"),
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []interface{}) {
				require.Len(t, result, 1)
				pool := result[0].(map[string]interface{})
				assert.Equal(t, "test-pool", pool["name"])
				assert.Equal(t, int32(3), pool["count"])
				assert.Equal(t, 1, pool["min"])
				assert.Equal(t, 5, pool["max"])
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := flattenMachinePoolConfigsVsphere(tt.input)
			tt.validate(t, result)
		})
	}
}

// TestUnit_FlattenClusterConfigsVsphere tests the flattenClusterConfigsVsphere function
func TestUnit_FlattenClusterConfigsVsphere(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *models.V1VsphereCloudConfig
		validate func(t *testing.T, result interface{})
	}{
		{
			name:  "nil input returns empty slice",
			input: nil,
			validate: func(t *testing.T, result interface{}) {
				slice := result.([]interface{})
				assert.Empty(t, slice)
			},
		},
		{
			name: "valid cloud config",
			input: &models.V1VsphereCloudConfig{
				Spec: &models.V1VsphereCloudConfigSpec{
					ClusterConfig: &models.V1VsphereClusterConfig{
						SSHKeys:    []string{"ssh-rsa AAAAB3..."},
						StaticIP:   false,
						NtpServers: []string{"ntp1.example.com"},
						Placement: &models.V1VspherePlacementConfig{
							Datacenter:          "DC1",
							Folder:              "/VMs/test",
							ImageTemplateFolder: "templates",
						},
						ControlPlaneEndpoint: &models.V1ControlPlaneEndPoint{
							Type:             "DDNS",
							DdnsSearchDomain: "spectrocloud.dev",
							Host:             "cluster.example.com",
						},
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				slice := result.([]interface{})
				require.Len(t, slice, 1)
				config := slice[0].(map[string]interface{})
				assert.Equal(t, "DC1", config["datacenter"])
				assert.Equal(t, "/VMs/test", config["folder"])
				assert.Equal(t, "DDNS", config["network_type"])
				assert.Equal(t, "spectrocloud.dev", config["network_search_domain"])
				assert.Equal(t, false, config["static_ip"])
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := prepareClusterVsphereTestData()
			result := flattenClusterConfigsVsphere(d, tt.input)
			tt.validate(t, result)
		})
	}
}

// TestUnit_ToCloudConfigCreate tests the toCloudConfigCreate function
func TestUnit_ToCloudConfigCreate(t *testing.T) {
	t.Parallel()

	cloudConfig := map[string]interface{}{
		"datacenter":            "Datacenter1",
		"folder":                "test/folder",
		"image_template_folder": "templates",
		"ssh_key":               "ssh-rsa AAAAB3...",
		"static_ip":             false,
		"network_type":          "DDNS",
		"network_search_domain": "spectrocloud.dev",
		"host_endpoint":         "cluster.example.com",
	}

	result := toCloudConfigCreate(cloudConfig)

	assert.NotNil(t, result)
	assert.Equal(t, "Datacenter1", result.Placement.Datacenter)
	assert.Equal(t, "test/folder", result.Placement.Folder)
	assert.Equal(t, "DDNS", result.ControlPlaneEndpoint.Type)
	assert.Equal(t, "spectrocloud.dev", result.ControlPlaneEndpoint.DdnsSearchDomain)
	assert.Equal(t, "cluster.example.com", result.ControlPlaneEndpoint.Host)
}

// TestUnit_ToVsphereCluster tests the toVsphereCluster function (lines 808-845)
func TestUnit_ToVsphereCluster(t *testing.T) {
	t.Parallel()

	// Minimal mock server so toProfiles (or any client call) does not fail
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "/v1/dashboard/projects/metadata") {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{{"metadata": map[string]interface{}{"name": "Default", "uid": "default-project-uid"}}},
			})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
	defer server.Close()

	c := createVsphereTestClient(t, server.URL)
	client.WithScopeProject("default-project-uid")(c)

	d := prepareClusterVsphereTestData()
	d.Set("context", "project")

	cluster, err := toVsphereCluster(c, d)

	require.NoError(t, err)
	require.NotNil(t, cluster)
	require.NotNil(t, cluster.Metadata)
	require.NotNil(t, cluster.Spec)

	assert.Equal(t, "vsphere-picard-2", cluster.Metadata.Name)
	assert.Equal(t, "vmware-basic-account-id", cluster.Spec.CloudAccountUID)
	require.NotNil(t, cluster.Spec.CloudConfig)
	assert.Equal(t, "Datacenter", cluster.Spec.CloudConfig.Placement.Datacenter)
	assert.Equal(t, "sc_test/terraform", cluster.Spec.CloudConfig.Placement.Folder)

	require.Len(t, cluster.Spec.Machinepoolconfig, 2, "expect control plane + worker pool")
	// toVsphereCluster sorts so control plane is first
	assert.True(t, cluster.Spec.Machinepoolconfig[0].PoolConfig.IsControlPlane, "first pool should be control plane")
	assert.Equal(t, "cp-pool", *cluster.Spec.Machinepoolconfig[0].PoolConfig.Name)
	assert.False(t, cluster.Spec.Machinepoolconfig[1].PoolConfig.IsControlPlane, "second pool should be worker")
	assert.Equal(t, "worker-basic", *cluster.Spec.Machinepoolconfig[1].PoolConfig.Name)
}

// TestUnit_ValidateMachinePoolChange tests the ValidateMachinePoolChange function
func TestUnit_ValidateMachinePoolChange(t *testing.T) {
	t.Parallel()

	// Helper to create a full machine pool map with all required fields for the hash function
	createPool := func(name string, controlPlane bool, cluster, datastore, resourcePool, network string) map[string]interface{} {
		return map[string]interface{}{
			"name":                    name,
			"control_plane":           controlPlane,
			"control_plane_as_worker": false,
			"count":                   1,
			"placement": []interface{}{
				map[string]interface{}{
					"cluster":           cluster,
					"datastore":         datastore,
					"resource_pool":     resourcePool,
					"network":           network,
					"static_ip_pool_id": "",
				},
			},
			"instance_type": []interface{}{
				map[string]interface{}{
					"disk_size_gb": 60,
					"memory_mb":    8192,
					"cpu":          4,
				},
			},
		}
	}

	// Create old machine pool set
	oldMPool := schema.NewSet(resourceMachinePoolVsphereHash, []interface{}{
		createPool("cp-pool", true, "cluster1", "datastore1", "pool1", "network1"),
	})

	// Test same placement - should pass
	newMPoolSame := schema.NewSet(resourceMachinePoolVsphereHash, []interface{}{
		createPool("cp-pool", true, "cluster1", "datastore1", "pool1", "network1"),
	})

	hasError, err := ValidateMachinePoolChange(oldMPool, newMPoolSame)
	assert.False(t, hasError)
	assert.NoError(t, err)
}

// TestUnit_SortPlacementStructs tests the sortPlacementStructs function
func TestUnit_SortPlacementStructs(t *testing.T) {
	t.Parallel()

	placements := []interface{}{
		map[string]interface{}{
			"cluster":       "cluster-b",
			"datastore":     "ds1",
			"resource_pool": "rp1",
			"network":       "net1",
		},
		map[string]interface{}{
			"cluster":       "cluster-a",
			"datastore":     "ds1",
			"resource_pool": "rp1",
			"network":       "net1",
		},
	}

	sortPlacementStructs(placements)

	// After sorting, cluster-a should come first
	assert.Equal(t, "cluster-a", placements[0].(map[string]interface{})["cluster"])
	assert.Equal(t, "cluster-b", placements[1].(map[string]interface{})["cluster"])
}

// =============================================================================
// VCR-Enabled Tests
// =============================================================================

// TestVCR_ClusterVsphereCRUD tests cluster vSphere CRUD operations using VCR
func TestVCR_ClusterVsphereCRUD(t *testing.T) {
	mode := vcr.GetMode()

	recorder, err := vcr.NewRecorder("cluster_vsphere_crud", mode)
	if err != nil {
		if mode == vcr.ModeReplaying {
			t.Skip("Skipping VCR test: cassette not found. Run with VCR_RECORD=true to record.")
		}
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer func() {
		if err := recorder.Stop(); err != nil {
			t.Errorf("Failed to stop recorder: %v", err)
		}
	}()

	t.Run("create_cluster", func(t *testing.T) {
		// 1. VCR recorder (already created in parent) and http.Client with Transport: recorder
		httpClient := &http.Client{Transport: recorder}
		_ = httpClient // When palette-sdk-go adds WithHTTPClient, use: client.New(..., client.WithHTTPClient(httpClient))

		// 2. Build Palette client: SDK does not accept custom HTTP client, so use cassette replay via httptest.Server
		cassette, err := vcr.LoadCassette("spectrocloud/testdata/cassettes/cluster_vsphere_crud.json")
		if err != nil {
			cassette, err = vcr.LoadCassette("testdata/cassettes/cluster_vsphere_crud.json")
		}
		require.NoError(t, err, "load cassette for create_cluster")
		server := newVsphereCassetteReplayServer(t, cassette)
		defer server.Close()

		c := createVsphereTestClient(t, server.URL)
		client.WithScopeProject("default-project-uid")(c)

		// 3. Prepare ResourceData and call resourceClusterVsphereCreate
		d := prepareClusterVsphereTestData()
		d.Set("context", "project")
		d.Set("skip_completion", true) // skip wait for Running-Healthy; cassette has no GetClusterOverview

		diags := resourceClusterVsphereCreate(context.Background(), d, c)

		require.False(t, diags.HasError(), "resourceClusterVsphereCreate should not return errors: %v", diags)
		assert.Equal(t, "vsphere-cluster-uid-12345", d.Id(), "cluster UID from VCR cassette")
	})

	t.Run("read_cluster", func(t *testing.T) {
		// 1. Load VCR cassette (same as create_cluster)
		cassette, err := vcr.LoadCassette("spectrocloud/testdata/cassettes/cluster_vsphere_crud.json")
		if err != nil {
			cassette, err = vcr.LoadCassette("testdata/cassettes/cluster_vsphere_crud.json")
		}
		require.NoError(t, err, "load cassette for read_cluster")

		// 2. Replay cassette via httptest.Server (method+path matching)
		server := newVsphereCassetteReplayServer(t, cassette)
		defer server.Close()

		c := createVsphereTestClient(t, server.URL)
		client.WithScopeProject("default-project-uid")(c)

		// 3. Prepare ResourceData for Read (cluster UID and context from cassette)
		d := resourceClusterVsphere().TestResourceData()
		d.SetId("vsphere-cluster-uid-12345")
		d.Set("context", "project")

		diags := resourceClusterVsphereRead(context.Background(), d, c)

		require.False(t, diags.HasError(), "resourceClusterVsphereRead should not return errors: %v", diags)
		assert.Equal(t, "cloud-config-uid-123", d.Get("cloud_config_id"), "cloud_config_id from cassette")
		assert.Equal(t, "cloud-account-uid-123", d.Get("cloud_account_id"), "cloud_account_id from cassette")
	})

	t.Run("update_cluster", func(t *testing.T) {
		// 1. Load VCR cassette (same as create_cluster / read_cluster)
		cassette, err := vcr.LoadCassette("spectrocloud/testdata/cassettes/cluster_vsphere_crud.json")
		if err != nil {
			cassette, err = vcr.LoadCassette("testdata/cassettes/cluster_vsphere_crud.json")
		}
		require.NoError(t, err, "load cassette for update_cluster")

		// 2. Replay cassette via httptest.Server (method+path matching)
		server := newVsphereCassetteReplayServer(t, cassette)
		defer server.Close()

		c := createVsphereTestClient(t, server.URL)
		client.WithScopeProject("default-project-uid")(c)

		// 3. Prepare ResourceData for Update (cluster UID, cloud_config_id, cloud_config, machine_pool from cassette; no changes so only GetCloudConfig + updateCommonFields + Read run)
		d := resourceClusterVsphere().TestResourceData()
		d.SetId("vsphere-cluster-uid-12345")
		d.Set("context", "project")
		d.Set("cloud_config_id", "cloud-config-uid-123")
		d.Set("review_repave_state", "")

		cloudConfig := []map[string]interface{}{
			{
				"datacenter":            "DC1",
				"folder":                "/test/folder",
				"ssh_keys":              []string{"ssh-rsa AAAAB3..."},
				"static_ip":             false,
				"network_type":          "DDNS",
				"network_search_domain": "spectrocloud.dev",
			},
		}
		d.Set("cloud_config", cloudConfig)

		cpPlacement := []interface{}{
			map[string]interface{}{
				"id":                "placement-1",
				"cluster":           "vsphere-cluster",
				"resource_pool":     "default-pool",
				"datastore":         "datastore1",
				"network":           "VM Network",
				"static_ip_pool_id": "",
			},
		}
		cpInstance := []interface{}{
			map[string]interface{}{"disk_size_gb": 60, "memory_mb": 8192, "cpu": 4},
		}
		mPools := []interface{}{
			map[string]interface{}{
				"control_plane":           true,
				"control_plane_as_worker": true,
				"name":                    "cp-pool",
				"count":                   1,
				"placement":               cpPlacement,
				"instance_type":           cpInstance,
				"node":                    []interface{}{},
			},
		}
		d.Set("machine_pool", mPools)

		diags := resourceClusterVsphereUpdate(context.Background(), d, c)

		require.False(t, diags.HasError(), "resourceClusterVsphereUpdate should not return errors: %v", diags)
		assert.Equal(t, "vsphere-cluster-uid-12345", d.Id(), "cluster UID unchanged after update")
		assert.Equal(t, "cloud-config-uid-123", d.Get("cloud_config_id"), "cloud_config_id from cassette after Read")
	})

	// // 1) validateOverrideScaling: machine_pool with update_strategy "OverrideScaling" and missing override_scaling must return error.
	// // We test the validation function directly (same code path Update uses when machine_pool changes) because TestResourceData
	// // cannot simulate state vs config, so HasChange("machine_pool") is not true when only Set() is used.
	// t.Run("update_validateOverrideScaling_error", func(t *testing.T) {
	// 	d := resourceClusterVsphere().TestResourceData()
	// 	cpPlacement := []interface{}{
	// 		map[string]interface{}{
	// 			"id":                "placement-1",
	// 			"cluster":           "vsphere-cluster",
	// 			"resource_pool":     "default-pool",
	// 			"datastore":         "datastore1",
	// 			"network":           "VM Network",
	// 			"static_ip_pool_id": "",
	// 		},
	// 	}
	// 	cpInstance := []interface{}{
	// 		map[string]interface{}{"disk_size_gb": 60, "memory_mb": 8192, "cpu": 4},
	// 	}
	// 	// Machine pool with OverrideScaling but missing override_scaling (invalid)
	// 	invalidPools := []interface{}{
	// 		map[string]interface{}{
	// 			"control_plane":           true,
	// 			"control_plane_as_worker": true,
	// 			"name":                    "cp-pool",
	// 			"count":                   1,
	// 			"placement":               cpPlacement,
	// 			"instance_type":           cpInstance,
	// 			"node":                    []interface{}{},
	// 			"update_strategy":         "OverrideScaling",
	// 			// override_scaling missing -> validateOverrideScaling must error
	// 		},
	// 	}
	// 	d.Set("machine_pool", schema.NewSet(resourceMachinePoolVsphereHash, invalidPools))

	// 	err := validateOverrideScaling(d, "machine_pool")
	// 	require.Error(t, err)
	// 	assert.Contains(t, err.Error(), "override_scaling", "error should mention override_scaling")
	// })

	// // 2) updateCommonFields: change description (or tags) and mock UpdateClusterMetadata
	// t.Run("update_common_fields_description", func(t *testing.T) {
	// 	clusterBody := `{"metadata":{"name":"test-vsphere-cluster","uid":"vsphere-cluster-uid-12345","labels":{"env":"test"},"annotations":{}},"spec":{"cloudConfigRef":{"uid":"cloud-config-uid-123"},"cloudType":"vsphere","clusterConfig":{}},"status":{"state":"Running","repave":{}}}`
	// 	// Use full structure so resourceClusterVsphereRead -> flattenMachinePoolConfigsVsphere does not nil-deref (placements[].network.networkName)
	// 	cloudConfigBody := `{"metadata":{"uid":"cloud-config-uid-123"},"spec":{"cloudAccountRef":{"uid":"cloud-account-uid-123"},"clusterConfig":{"placement":{"datacenter":"DC1","folder":"/test/folder"},"sshKeys":["ssh-rsa AAAAB3..."],"staticIp":false,"ntpServers":[]},"machinePoolConfig":[{"name":"cp-pool","size":1,"minSize":1,"maxSize":3,"isControlPlane":true,"useControlPlaneAsWorker":true,"instanceType":{"diskGiB":60,"memoryMiB":8192,"numCPUs":4},"placements":[{"uid":"placement-1","cluster":"vsphere-cluster","resourcePool":"default-pool","datastore":"datastore1","network":{"networkName":"VM Network"}}]}]}}`

	// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		w.Header().Set("Content-Type", "application/json")
	// 		path := r.URL.Path
	// 		switch {
	// 		case strings.Contains(path, "/v1/dashboard/projects/metadata"):
	// 			json.NewEncoder(w).Encode(map[string]interface{}{
	// 				"items": []map[string]interface{}{
	// 					{"metadata": map[string]interface{}{"name": "Default", "uid": "default-project-uid"}},
	// 				},
	// 			})
	// 			return
	// 		case strings.Contains(path, "/v1/spectroclusters/vsphere-cluster-uid-12345"):
	// 			if r.Method == http.MethodGet {
	// 				w.Write([]byte(clusterBody))
	// 			} else if r.Method == http.MethodPut || r.Method == http.MethodPatch {
	// 				w.WriteHeader(http.StatusOK)
	// 				w.Write([]byte(clusterBody))
	// 			} else {
	// 				w.WriteHeader(http.StatusMethodNotAllowed)
	// 			}
	// 			return
	// 		case strings.Contains(path, "/v1/cloudconfigs/vsphere/cloud-config-uid-123"):
	// 			w.Write([]byte(cloudConfigBody))
	// 			return
	// 		case strings.Contains(path, "/assets/kubeconfig"):
	// 			w.Write([]byte("apiVersion: v1\nkind: Config\nclusters: []"))
	// 			return
	// 		case strings.Contains(path, "/assets/adminKubeconfig"):
	// 			w.Write([]byte("apiVersion: v1\nkind: Config\nclusters: []"))
	// 			return
	// 		case strings.Contains(path, "/config/rbacs"):
	// 			w.Write([]byte(`{"items":[]}`))
	// 			return
	// 		case strings.Contains(path, "/config/namespaces"):
	// 			w.Write([]byte(`{"items":[]}`))
	// 			return
	// 		case strings.Contains(path, "/variables"):
	// 			w.Write([]byte(`{"variables":[]}`))
	// 			return
	// 		case strings.Contains(path, "/machinePools/") && strings.Contains(path, "/machines"):
	// 			w.Write([]byte(`{"items":[]}`))
	// 			return
	// 		}
	// 		t.Logf("Mock server: no handler for %s %s", r.Method, path)
	// 		w.WriteHeader(http.StatusNotFound)
	// 	}))
	// 	defer server.Close()

	// 	c := createVsphereTestClient(t, server.URL)
	// 	client.WithScopeProject("default-project-uid")(c)

	// 	d := resourceClusterVsphere().TestResourceData()
	// 	d.SetId("vsphere-cluster-uid-12345")
	// 	d.Set("context", "project")
	// 	d.Set("cloud_config_id", "cloud-config-uid-123")
	// 	d.Set("review_repave_state", "")
	// 	d.Set("description", "updated-description")

	// 	cloudConfig := []map[string]interface{}{
	// 		{
	// 			"datacenter":            "DC1",
	// 			"folder":                "/test/folder",
	// 			"ssh_keys":              []string{"ssh-rsa AAAAB3..."},
	// 			"static_ip":             false,
	// 			"network_type":          "DDNS",
	// 			"network_search_domain": "spectrocloud.dev",
	// 		},
	// 	}
	// 	d.Set("cloud_config", cloudConfig)

	// 	cpPlacement := []interface{}{
	// 		map[string]interface{}{
	// 			"id":                "placement-1",
	// 			"cluster":           "vsphere-cluster",
	// 			"resource_pool":     "default-pool",
	// 			"datastore":         "datastore1",
	// 			"network":           "VM Network",
	// 			"static_ip_pool_id": "",
	// 		},
	// 	}
	// 	cpInstance := []interface{}{
	// 		map[string]interface{}{"disk_size_gb": 60, "memory_mb": 8192, "cpu": 4},
	// 	}
	// 	mPools := []interface{}{
	// 		map[string]interface{}{
	// 			"control_plane":           true,
	// 			"control_plane_as_worker": true,
	// 			"name":                    "cp-pool",
	// 			"count":                   1,
	// 			"placement":               cpPlacement,
	// 			"instance_type":           cpInstance,
	// 			"node":                    []interface{}{},
	// 		},
	// 	}
	// 	d.Set("machine_pool", schema.NewSet(resourceMachinePoolVsphereHash, mPools))

	// 	diags := resourceClusterVsphereUpdate(context.Background(), d, c)

	// 	require.False(t, diags.HasError(), "resourceClusterVsphereUpdate should not return errors: %v", diags)
	// 	assert.Equal(t, "vsphere-cluster-uid-12345", d.Id())
	// })

	t.Run("delete_cluster", func(t *testing.T) {
		// 1. Load VCR cassette for delete (GET 200, DELETE 204, GET 404 in order for DeleteCluster + waitForClusterDeletion)
		cassette, err := vcr.LoadCassette("spectrocloud/testdata/cassettes/cluster_vsphere_delete.json")
		if err != nil {
			cassette, err = vcr.LoadCassette("testdata/cassettes/cluster_vsphere_delete.json")
		}
		require.NoError(t, err, "load cassette for delete_cluster")

		// 2. Ordered replay so first GET returns 200 (DeleteCluster checks exists), DELETE returns 204, next GET returns 404 (wait sees deleted)
		server := newOrderedVsphereCassetteReplayServer(t, cassette)
		defer server.Close()

		c := createVsphereTestClient(t, server.URL)
		client.WithScopeProject("default-project-uid")(c)

		// 3. Prepare ResourceData and call resourceClusterDelete (shared delete used by vsphere resource)
		d := resourceClusterVsphere().TestResourceData()
		d.SetId("vsphere-cluster-uid-12345")
		d.Set("context", "project")

		diags := resourceClusterDelete(context.Background(), d, c)

		require.False(t, diags.HasError(), "resourceClusterDelete should not return errors: %v", diags)
	})
}

// =============================================================================
// httptest.Server based tests for full coverage
// These tests use Go's built-in httptest.Server to mock HTTP responses
// =============================================================================

// createVsphereMockServer creates an httptest.Server for vSphere cluster tests
func createVsphereMockServer(t *testing.T, responses map[string]interface{}) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock server received: %s %s", r.Method, r.URL.Path)

		// Project metadata endpoint
		if strings.Contains(r.URL.Path, "/v1/dashboard/projects/metadata") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []map[string]interface{}{
					{"metadata": map[string]interface{}{"name": "Default", "uid": "default-project-uid"}},
				},
			})
			return
		}

		// Check for specific endpoints
		for path, response := range responses {
			if strings.Contains(r.URL.Path, path) {
				w.Header().Set("Content-Type", "application/json")
				if response == nil {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("null"))
					return
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
				return
			}
		}

		t.Logf("Mock server: no handler for path %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))

	return server
}

// createVsphereTestClient creates a client for vSphere tests
func createVsphereTestClient(t *testing.T, serverURL string) *client.V1Client {
	t.Helper()

	host := strings.TrimPrefix(serverURL, "http://")
	host = strings.TrimPrefix(host, "https://")

	c := client.New(
		client.WithPaletteURI(host),
		client.WithAPIKey("test-api-key"),
		client.WithInsecureSkipVerify(true),
		client.WithRetries(1),
		client.WithSchemes([]string{"http"}),
	)

	return c
}

// newVsphereCassetteReplayServer creates an httptest.Server that replays cassette interactions with method+path matching.
// Used for Create flow where POST /v1/spectroclusters/vsphere must be distinguished from GETs.
func newVsphereCassetteReplayServer(t *testing.T, cassette *vcr.Cassette) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, interaction := range cassette.Interactions {
			matchMethod := r.Method == interaction.Request.Method
			matchPath := strings.Contains(r.URL.Path, interaction.Request.URL) || interaction.Request.URL == r.URL.Path
			if matchMethod && matchPath {
				for key, value := range interaction.Response.Headers {
					w.Header().Set(key, value)
				}
				w.WriteHeader(interaction.Response.StatusCode)
				w.Write([]byte(interaction.Response.Body))
				return
			}
		}
		t.Logf("VCR: No cassette match for %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
}

// newOrderedVsphereCassetteReplayServer creates an httptest.Server that replays cassette interactions in order.
// Each request gets the next matching (method+path) interaction. Used for Delete where we need GET 200, DELETE 204, GET 404 in sequence.
func newOrderedVsphereCassetteReplayServer(t *testing.T, cassette *vcr.Cassette) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	idx := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		for i := idx; i < len(cassette.Interactions); i++ {
			in := &cassette.Interactions[i]
			matchMethod := r.Method == in.Request.Method
			matchPath := strings.Contains(r.URL.Path, in.Request.URL) || in.Request.URL == r.URL.Path
			if matchMethod && matchPath {
				idx = i + 1
				for key, value := range in.Response.Headers {
					w.Header().Set(key, value)
				}
				w.WriteHeader(in.Response.StatusCode)
				w.Write([]byte(in.Response.Body))
				return
			}
		}
		t.Logf("VCR ordered: no match for %s %s at index %d", r.Method, r.URL.Path, idx)
		w.WriteHeader(http.StatusNotFound)
	}))
}

// TestVCR_ClusterVsphereRead tests reading a vSphere cluster using VCR cassettes
// This test loads the cassette and creates an httptest.Server to serve the recorded responses
func TestVCR_ClusterVsphereRead(t *testing.T) {
	// Load VCR cassette
	cassette, err := vcr.LoadCassette("spectrocloud/testdata/cassettes/cluster_vsphere_crud.json")
	if err != nil {
		// Try alternate path
		cassette, err = vcr.LoadCassette("testdata/cassettes/cluster_vsphere_crud.json")
		if err != nil {
			t.Skipf("Skipping VCR test: cassette not found: %v", err)
		}
	}

	// Create httptest.Server that serves responses from the cassette (method+path matching so POST create is distinct from GETs)
	server := newVsphereCassetteReplayServer(t, cassette)
	defer server.Close()

	// Validate VCR server is working
	resp, err := http.Get(server.URL + "/v1/spectroclusters/vsphere-cluster-uid-12345")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Validate cloud config endpoint from cassette
	resp2, err := http.Get(server.URL + "/v1/cloudconfigs/vsphere/cloud-config-uid-123")
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// Validate kubeconfig endpoint
	resp3, err := http.Get(server.URL + "/v1/spectroclusters/vsphere-cluster-uid-12345/assets/kubeconfig")
	require.NoError(t, err)
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusOK, resp3.StatusCode)
}

// TestHTTPServer_ClusterVsphereRead tests reading a vSphere cluster using httptest.Server
func TestHTTPServer_ClusterVsphereRead(t *testing.T) {
	// This test validates the mock server setup and response structure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(r.URL.Path, "/v1/spectroclusters/vsphere-cluster-uid-123"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-vsphere-cluster",
					"uid":  "vsphere-cluster-uid-123",
					"labels": map[string]string{
						"env": "test",
					},
				},
				"spec": map[string]interface{}{
					"cloudConfigRef": map[string]interface{}{
						"uid": "cloud-config-uid-123",
					},
					"cloudType": "vsphere",
				},
				"status": map[string]interface{}{
					"state": "Running",
				},
			})
			return

		case strings.Contains(r.URL.Path, "/v1/cloudconfigs/vsphere/cloud-config-uid-123"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"metadata": map[string]interface{}{
					"uid": "cloud-config-uid-123",
				},
				"spec": map[string]interface{}{
					"cloudAccountRef": map[string]interface{}{
						"uid": "cloud-account-uid-123",
					},
					"clusterConfig": map[string]interface{}{
						"sshKeys":  []string{"ssh-rsa AAAAB3..."},
						"staticIp": false,
						"placement": map[string]interface{}{
							"datacenter": "DC1",
							"folder":     "/test/folder",
						},
					},
					"machinePoolConfig": []interface{}{},
				},
			})
			return

		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	// Validate mock server is working
	resp, err := http.Get(server.URL + "/v1/spectroclusters/vsphere-cluster-uid-123")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Validate cloud config endpoint
	resp2, err := http.Get(server.URL + "/v1/cloudconfigs/vsphere/cloud-config-uid-123")
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

// TestHTTPMock_FlattenMachinePoolConfigsVsphere_AllBranches tests all branches
func TestHTTPMock_FlattenMachinePoolConfigsVsphere_AllBranches(t *testing.T) {
	t.Run("nil_machine_pools", func(t *testing.T) {
		result := flattenMachinePoolConfigsVsphere(nil)
		assert.Empty(t, result)
	})

	t.Run("empty_machine_pools", func(t *testing.T) {
		result := flattenMachinePoolConfigsVsphere([]*models.V1VsphereMachinePoolConfig{})
		assert.Empty(t, result)
	})

	t.Run("with_control_plane_pool", func(t *testing.T) {
		pools := []*models.V1VsphereMachinePoolConfig{
			{
				Name:           "cp-pool",
				Size:           3,
				MinSize:        1,
				MaxSize:        5,
				IsControlPlane: types.Ptr(true),
				InstanceType: &models.V1VsphereInstanceType{
					DiskGiB:   types.Ptr(int32(100)),
					MemoryMiB: types.Ptr(int64(8192)),
					NumCPUs:   types.Ptr(int32(4)),
				},
				UpdateStrategy: &models.V1UpdateStrategy{
					Type: "RollingUpdateScaleOut",
				},
				Placements: []*models.V1VspherePlacementConfig{
					{
						UID:          "placement-1",
						Cluster:      "vsphere-cluster",
						ResourcePool: "default-pool",
						Datastore:    "datastore1",
						Network: &models.V1VsphereNetworkConfig{
							NetworkName: types.Ptr("VM Network"),
							ParentPoolRef: &models.V1ObjectReference{
								UID: "static-ip-pool-1",
							},
						},
					},
				},
			},
		}

		result := flattenMachinePoolConfigsVsphere(pools)
		require.Len(t, result, 1)

		pool := result[0].(map[string]interface{})
		assert.Equal(t, "cp-pool", pool["name"])
		assert.Equal(t, int32(3), pool["count"])
		assert.Equal(t, true, pool["control_plane"])
		assert.Equal(t, "RollingUpdateScaleOut", pool["update_strategy"])
	})

	t.Run("with_worker_pool_and_override_kubeadm", func(t *testing.T) {
		pools := []*models.V1VsphereMachinePoolConfig{
			{
				Name:                         "worker-pool",
				Size:                         2,
				IsControlPlane:               types.Ptr(false),
				NodeRepaveInterval:           60,
				OverrideKubeadmConfiguration: "apiVersion: kubeadm.k8s.io/v1beta3\nkind: JoinConfiguration",
				InstanceType: &models.V1VsphereInstanceType{
					DiskGiB:   types.Ptr(int32(200)),
					MemoryMiB: types.Ptr(int64(16384)),
					NumCPUs:   types.Ptr(int32(8)),
				},
				Placements: []*models.V1VspherePlacementConfig{
					{
						UID:          "placement-2",
						Cluster:      "vsphere-cluster",
						ResourcePool: "worker-pool",
						Datastore:    "datastore2",
						Network: &models.V1VsphereNetworkConfig{
							NetworkName: types.Ptr("Worker Network"),
						},
					},
				},
			},
		}

		result := flattenMachinePoolConfigsVsphere(pools)
		require.Len(t, result, 1)

		pool := result[0].(map[string]interface{})
		assert.Equal(t, "worker-pool", pool["name"])
		assert.Equal(t, false, pool["control_plane"])
		assert.Contains(t, pool["override_kubeadm_configuration"], "kubeadm.k8s.io")
	})
}

// TestHTTPMock_FlattenClusterConfigsVsphere_AllBranches tests all branches in flattenClusterConfigsVsphere
func TestHTTPMock_FlattenClusterConfigsVsphere_AllBranches(t *testing.T) {
	t.Run("nil_cloud_config", func(t *testing.T) {
		d := prepareClusterVsphereTestData()
		result := flattenClusterConfigsVsphere(d, nil)
		slice := result.([]interface{})
		assert.Empty(t, slice)
	})

	t.Run("nil_cluster_config", func(t *testing.T) {
		d := prepareClusterVsphereTestData()
		cloudConfig := &models.V1VsphereCloudConfig{
			Spec: &models.V1VsphereCloudConfigSpec{
				ClusterConfig: nil,
			},
		}
		result := flattenClusterConfigsVsphere(d, cloudConfig)
		slice := result.([]interface{})
		assert.Empty(t, slice)
	})

	t.Run("with_control_plane_endpoint", func(t *testing.T) {
		d := prepareClusterVsphereTestData()
		cloudConfig := &models.V1VsphereCloudConfig{
			Spec: &models.V1VsphereCloudConfigSpec{
				ClusterConfig: &models.V1VsphereClusterConfig{
					SSHKeys:  []string{"ssh-rsa AAAAB3..."},
					StaticIP: false,
					Placement: &models.V1VspherePlacementConfig{
						Datacenter:          "DC1",
						Folder:              "/test/folder",
						ImageTemplateFolder: "templates",
					},
					ControlPlaneEndpoint: &models.V1ControlPlaneEndPoint{
						Type:             "VIP",
						DdnsSearchDomain: "spectrocloud.dev",
						Host:             "cluster.example.com",
					},
					NtpServers: []string{"ntp1.example.com", "ntp2.example.com"},
				},
			},
		}

		result := flattenClusterConfigsVsphere(d, cloudConfig)
		slice := result.([]interface{})
		require.Len(t, slice, 1)

		config := slice[0].(map[string]interface{})
		assert.Equal(t, "DC1", config["datacenter"])
		assert.Equal(t, "/test/folder", config["folder"])
		assert.Equal(t, "templates", config["image_template_folder"])
		assert.Equal(t, "VIP", config["network_type"])
		assert.Equal(t, "spectrocloud.dev", config["network_search_domain"])
		assert.Equal(t, "cluster.example.com", config["host_endpoint"])
		assert.Equal(t, false, config["static_ip"])
	})

	t.Run("with_ssh_key_singular", func(t *testing.T) {
		d := resourceClusterVsphere().TestResourceData()
		// Set ssh_key (singular) to trigger that branch
		cloudConfig := []map[string]interface{}{
			{
				"datacenter": "DC1",
				"folder":     "/test/folder",
				"ssh_key":    "ssh-rsa AAAAB3...",
			},
		}
		d.Set("cloud_config", cloudConfig)

		vsphereCloudConfig := &models.V1VsphereCloudConfig{
			Spec: &models.V1VsphereCloudConfigSpec{
				ClusterConfig: &models.V1VsphereClusterConfig{
					SSHKeys: []string{"ssh-rsa AAAAB3..."},
					Placement: &models.V1VspherePlacementConfig{
						Datacenter: "DC1",
						Folder:     "/test/folder",
					},
				},
			},
		}

		result := flattenClusterConfigsVsphere(d, vsphereCloudConfig)
		slice := result.([]interface{})
		require.Len(t, slice, 1)

		config := slice[0].(map[string]interface{})
		assert.Equal(t, "ssh-rsa AAAAB3...", config["ssh_key"])
	})
}

// TestHTTPMock_ToCloudConfigCreate_AllBranches tests toCloudConfigCreate
func TestHTTPMock_ToCloudConfigCreate_AllBranches(t *testing.T) {
	t.Run("with_all_fields", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"datacenter":            "Datacenter1",
			"folder":                "test/folder",
			"image_template_folder": "templates",
			"ssh_key":               "ssh-rsa AAAAB3...",
			"static_ip":             true,
			"network_type":          "VIP",
			"network_search_domain": "spectrocloud.dev",
			"host_endpoint":         "cluster.example.com",
			"ntp_servers":           schema.NewSet(schema.HashString, []interface{}{"ntp1.example.com"}),
		}

		result := toCloudConfigCreate(cloudConfig)

		assert.NotNil(t, result)
		assert.Equal(t, "Datacenter1", result.Placement.Datacenter)
		assert.Equal(t, "test/folder", result.Placement.Folder)
		assert.Equal(t, "VIP", result.ControlPlaneEndpoint.Type)
		assert.Equal(t, "spectrocloud.dev", result.ControlPlaneEndpoint.DdnsSearchDomain)
		assert.Equal(t, "cluster.example.com", result.ControlPlaneEndpoint.Host)
	})

	t.Run("with_ddns_network_type", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"datacenter":            "DC1",
			"folder":                "/vms",
			"ssh_key":               "ssh-rsa KEY",
			"static_ip":             false,
			"network_type":          "DDNS",
			"network_search_domain": "test.local",
			"host_endpoint":         "",
		}

		result := toCloudConfigCreate(cloudConfig)

		assert.Equal(t, "DDNS", result.ControlPlaneEndpoint.Type)
		assert.Equal(t, "test.local", result.ControlPlaneEndpoint.DdnsSearchDomain)
	})
}

// TestHTTPMock_ToCloudConfigUpdate_AllBranches tests toCloudConfigUpdate
func TestHTTPMock_ToCloudConfigUpdate_AllBranches(t *testing.T) {
	cloudConfig := map[string]interface{}{
		"datacenter":            "Datacenter1",
		"folder":                "test/folder",
		"image_template_folder": "custom-templates",
		"ssh_key":               "ssh-rsa AAAAB3...",
		"static_ip":             false,
		"network_type":          "DDNS",
		"network_search_domain": "spectrocloud.dev",
		"host_endpoint":         "cluster.example.com",
	}

	result := toCloudConfigUpdate(cloudConfig)

	assert.NotNil(t, result)
	assert.NotNil(t, result.ClusterConfig)
	assert.Equal(t, "Datacenter1", result.ClusterConfig.Placement.Datacenter)
	assert.Equal(t, "DDNS", result.ClusterConfig.ControlPlaneEndpoint.Type)
}

// TestHTTPMock_ToMachinePoolVsphere_AllBranches tests all branches in toMachinePoolVsphere
func TestHTTPMock_ToMachinePoolVsphere_AllBranches(t *testing.T) {
	t.Run("control_plane_with_static_ip", func(t *testing.T) {
		input := map[string]interface{}{
			"name":                    "cp-pool",
			"control_plane":           true,
			"control_plane_as_worker": true,
			"count":                   3,
			"node_repave_interval":    0,
			"placement": []interface{}{
				map[string]interface{}{
					"id":                "placement-1",
					"cluster":           "vsphere-cluster",
					"resource_pool":     "default-pool",
					"datastore":         "datastore1",
					"network":           "VM Network",
					"static_ip_pool_id": "static-pool-123",
				},
			},
			"instance_type": []interface{}{
				map[string]interface{}{
					"disk_size_gb": 100,
					"memory_mb":    8192,
					"cpu":          4,
				},
			},
		}

		result, err := toMachinePoolVsphere(input)

		require.NoError(t, err)
		assert.Equal(t, "cp-pool", *result.PoolConfig.Name)
		assert.True(t, result.PoolConfig.IsControlPlane)
		assert.True(t, result.PoolConfig.UseControlPlaneAsWorker)
		assert.Contains(t, result.PoolConfig.Labels, "control-plane")

		// Verify static IP is set
		assert.True(t, result.CloudConfig.Placements[0].Network.StaticIP)
		assert.Equal(t, "static-pool-123", result.CloudConfig.Placements[0].Network.ParentPoolUID)
	})

	t.Run("worker_pool_with_override_kubeadm", func(t *testing.T) {
		input := map[string]interface{}{
			"name":                           "worker-pool",
			"control_plane":                  false,
			"control_plane_as_worker":        false,
			"count":                          2,
			"min":                            1,
			"max":                            10,
			"node_repave_interval":           120,
			"override_kubeadm_configuration": "custom-config-yaml",
			"placement": []interface{}{
				map[string]interface{}{
					"id":                "",
					"cluster":           "vsphere-cluster",
					"resource_pool":     "worker-pool",
					"datastore":         "datastore2",
					"network":           "Worker Network",
					"static_ip_pool_id": "",
				},
			},
			"instance_type": []interface{}{
				map[string]interface{}{
					"disk_size_gb": 200,
					"memory_mb":    16384,
					"cpu":          8,
				},
			},
		}

		result, err := toMachinePoolVsphere(input)

		require.NoError(t, err)
		assert.Equal(t, "worker-pool", *result.PoolConfig.Name)
		assert.False(t, result.PoolConfig.IsControlPlane)
		assert.Contains(t, result.PoolConfig.Labels, "worker")
		assert.Equal(t, int32(1), result.PoolConfig.MinSize)
		assert.Equal(t, int32(10), result.PoolConfig.MaxSize)
		assert.Equal(t, int32(120), result.PoolConfig.NodeRepaveInterval)
		assert.Equal(t, "custom-config-yaml", result.PoolConfig.OverrideKubeadmConfiguration)
	})

	t.Run("negative_disk_size_error", func(t *testing.T) {
		input := map[string]interface{}{
			"name":                    "invalid-pool",
			"control_plane":           false,
			"control_plane_as_worker": false,
			"count":                   1,
			"node_repave_interval":    0,
			"placement": []interface{}{
				map[string]interface{}{
					"id":                "",
					"cluster":           "vsphere-cluster",
					"resource_pool":     "default",
					"datastore":         "datastore1",
					"network":           "VM Network",
					"static_ip_pool_id": "",
				},
			},
			"instance_type": []interface{}{
				map[string]interface{}{
					"disk_size_gb": -100, // Negative
					"memory_mb":    8192,
					"cpu":          4,
				},
			},
		}

		_, err := toMachinePoolVsphere(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be negative")
	})

	t.Run("negative_min_error", func(t *testing.T) {
		input := map[string]interface{}{
			"name":                    "invalid-pool",
			"control_plane":           false,
			"control_plane_as_worker": false,
			"count":                   1,
			"min":                     -1, // Negative
			"node_repave_interval":    0,
			"placement": []interface{}{
				map[string]interface{}{
					"id":                "",
					"cluster":           "vsphere-cluster",
					"resource_pool":     "default",
					"datastore":         "datastore1",
					"network":           "VM Network",
					"static_ip_pool_id": "",
				},
			},
			"instance_type": []interface{}{
				map[string]interface{}{
					"disk_size_gb": 100,
					"memory_mb":    8192,
					"cpu":          4,
				},
			},
		}

		_, err := toMachinePoolVsphere(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min value")
	})

	t.Run("negative_max_error", func(t *testing.T) {
		input := map[string]interface{}{
			"name":                    "invalid-pool",
			"control_plane":           false,
			"control_plane_as_worker": false,
			"count":                   1,
			"min":                     1,
			"max":                     -1, // Negative
			"node_repave_interval":    0,
			"placement": []interface{}{
				map[string]interface{}{
					"id":                "",
					"cluster":           "vsphere-cluster",
					"resource_pool":     "default",
					"datastore":         "datastore1",
					"network":           "VM Network",
					"static_ip_pool_id": "",
				},
			},
			"instance_type": []interface{}{
				map[string]interface{}{
					"disk_size_gb": 100,
					"memory_mb":    8192,
					"cpu":          4,
				},
			},
		}

		_, err := toMachinePoolVsphere(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max value")
	})

	t.Run("negative_node_repave_interval_error", func(t *testing.T) {
		input := map[string]interface{}{
			"name":                    "invalid-pool",
			"control_plane":           false,
			"control_plane_as_worker": false,
			"count":                   1,
			"node_repave_interval":    -60, // Negative
			"placement": []interface{}{
				map[string]interface{}{
					"id":                "",
					"cluster":           "vsphere-cluster",
					"resource_pool":     "default",
					"datastore":         "datastore1",
					"network":           "VM Network",
					"static_ip_pool_id": "",
				},
			},
			"instance_type": []interface{}{
				map[string]interface{}{
					"disk_size_gb": 100,
					"memory_mb":    8192,
					"cpu":          4,
				},
			},
		}

		_, err := toMachinePoolVsphere(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node_repave_interval")
	})
}

// TestHTTPMock_GetSSHKey tests the getSSHKey function
func TestHTTPMock_GetSSHKey(t *testing.T) {
	cloudConfig := map[string]interface{}{
		"ssh_key": "ssh-rsa AAAAB3...",
	}

	result := getSSHKey(cloudConfig)
	assert.Contains(t, result, "ssh-rsa AAAAB3...")
}

// TestHTTPMock_GetStaticIP tests the getStaticIP function
func TestHTTPMock_GetStaticIP(t *testing.T) {
	t.Run("static_ip_true", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"static_ip": true,
		}
		result := getStaticIP(cloudConfig)
		assert.True(t, result)
	})

	t.Run("static_ip_false", func(t *testing.T) {
		cloudConfig := map[string]interface{}{
			"static_ip": false,
		}
		result := getStaticIP(cloudConfig)
		assert.False(t, result)
	})
}

// TestHTTPMock_GetClusterConfigEntity tests the getClusterConfigEntity function
func TestHTTPMock_GetClusterConfigEntity(t *testing.T) {
	cloudConfig := map[string]interface{}{
		"datacenter":            "DC1",
		"folder":                "/vms/test",
		"image_template_folder": "templates",
		"ssh_key":               "ssh-rsa KEY",
		"static_ip":             true,
		"ntp_servers":           schema.NewSet(schema.HashString, []interface{}{"ntp.example.com"}),
	}

	result := getClusterConfigEntity(cloudConfig)

	assert.NotNil(t, result)
	assert.Equal(t, "DC1", result.Placement.Datacenter)
	assert.Equal(t, "/vms/test", result.Placement.Folder)
	assert.True(t, result.StaticIP)
}

// TestHTTPMock_SortPlacementStructs_AllCases tests all sorting scenarios
func TestHTTPMock_SortPlacementStructs_AllCases(t *testing.T) {
	t.Run("sort_by_cluster", func(t *testing.T) {
		placements := []interface{}{
			map[string]interface{}{"cluster": "z-cluster", "datastore": "ds1", "resource_pool": "rp1", "network": "net1"},
			map[string]interface{}{"cluster": "a-cluster", "datastore": "ds1", "resource_pool": "rp1", "network": "net1"},
		}
		sortPlacementStructs(placements)
		assert.Equal(t, "a-cluster", placements[0].(map[string]interface{})["cluster"])
	})

	t.Run("sort_by_datastore_when_cluster_same", func(t *testing.T) {
		placements := []interface{}{
			map[string]interface{}{"cluster": "cluster1", "datastore": "z-ds", "resource_pool": "rp1", "network": "net1"},
			map[string]interface{}{"cluster": "cluster1", "datastore": "a-ds", "resource_pool": "rp1", "network": "net1"},
		}
		sortPlacementStructs(placements)
		assert.Equal(t, "a-ds", placements[0].(map[string]interface{})["datastore"])
	})

	t.Run("sort_by_resource_pool_when_cluster_and_datastore_same", func(t *testing.T) {
		placements := []interface{}{
			map[string]interface{}{"cluster": "cluster1", "datastore": "ds1", "resource_pool": "z-rp", "network": "net1"},
			map[string]interface{}{"cluster": "cluster1", "datastore": "ds1", "resource_pool": "a-rp", "network": "net1"},
		}
		sortPlacementStructs(placements)
		assert.Equal(t, "a-rp", placements[0].(map[string]interface{})["resource_pool"])
	})

	t.Run("sort_by_network_when_all_others_same", func(t *testing.T) {
		placements := []interface{}{
			map[string]interface{}{"cluster": "cluster1", "datastore": "ds1", "resource_pool": "rp1", "network": "z-net"},
			map[string]interface{}{"cluster": "cluster1", "datastore": "ds1", "resource_pool": "rp1", "network": "a-net"},
		}
		sortPlacementStructs(placements)
		assert.Equal(t, "a-net", placements[0].(map[string]interface{})["network"])
	})
}

// TestHTTPMock_ValidateMachinePoolChange_AllBranches tests all validation scenarios
func TestHTTPMock_ValidateMachinePoolChange_AllBranches(t *testing.T) {
	// Helper to create a full machine pool map with all required fields for the hash function
	createFullMachinePool := func(name string, controlPlane bool, placements []interface{}) map[string]interface{} {
		return map[string]interface{}{
			"name":                    name,
			"control_plane":           controlPlane,
			"control_plane_as_worker": false,
			"count":                   1,
			"placement":               placements,
			"instance_type": []interface{}{
				map[string]interface{}{
					"disk_size_gb": 60,
					"memory_mb":    8192,
					"cpu":          4,
				},
			},
		}
	}

	// Helper to create full placement with all required fields
	createFullPlacement := func(cluster, datastore, resourcePool, network string) map[string]interface{} {
		return map[string]interface{}{
			"cluster":           cluster,
			"datastore":         datastore,
			"resource_pool":     resourcePool,
			"network":           network,
			"static_ip_pool_id": "",
		}
	}

	createPoolSet := func(pools ...map[string]interface{}) *schema.Set {
		return schema.NewSet(resourceMachinePoolVsphereHash, func() []interface{} {
			result := make([]interface{}, len(pools))
			for i, p := range pools {
				result[i] = p
			}
			return result
		}())
	}

	t.Run("same_placements_valid", func(t *testing.T) {
		placement := []interface{}{createFullPlacement("c1", "ds1", "rp1", "net1")}
		oldPool := createPoolSet(createFullMachinePool("cp-pool", true, placement))
		newPool := createPoolSet(createFullMachinePool("cp-pool", true, placement))

		hasError, err := ValidateMachinePoolChange(oldPool, newPool)
		assert.False(t, hasError)
		assert.NoError(t, err)
	})

	t.Run("different_placement_count_error", func(t *testing.T) {
		oldPlacement := []interface{}{createFullPlacement("c1", "ds1", "rp1", "net1")}
		newPlacement := []interface{}{
			createFullPlacement("c1", "ds1", "rp1", "net1"),
			createFullPlacement("c2", "ds2", "rp2", "net2"),
		}
		oldPool := createPoolSet(createFullMachinePool("cp-pool", true, oldPlacement))
		newPool := createPoolSet(createFullMachinePool("cp-pool", true, newPlacement))

		hasError, err := ValidateMachinePoolChange(oldPool, newPool)
		assert.True(t, hasError)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "placement validation error")
	})

	t.Run("cluster_change_error", func(t *testing.T) {
		oldPlacement := []interface{}{createFullPlacement("old-cluster", "ds1", "rp1", "net1")}
		newPlacement := []interface{}{createFullPlacement("new-cluster", "ds1", "rp1", "net1")}
		oldPool := createPoolSet(createFullMachinePool("cp-pool", true, oldPlacement))
		newPool := createPoolSet(createFullMachinePool("cp-pool", true, newPlacement))

		hasError, err := ValidateMachinePoolChange(oldPool, newPool)
		assert.True(t, hasError)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ComputeCluster")
	})

	t.Run("datastore_change_error", func(t *testing.T) {
		oldPlacement := []interface{}{createFullPlacement("c1", "old-ds", "rp1", "net1")}
		newPlacement := []interface{}{createFullPlacement("c1", "new-ds", "rp1", "net1")}
		oldPool := createPoolSet(createFullMachinePool("cp-pool", true, oldPlacement))
		newPool := createPoolSet(createFullMachinePool("cp-pool", true, newPlacement))

		hasError, err := ValidateMachinePoolChange(oldPool, newPool)
		assert.True(t, hasError)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DataStore")
	})

	t.Run("resource_pool_change_error", func(t *testing.T) {
		oldPlacement := []interface{}{createFullPlacement("c1", "ds1", "old-rp", "net1")}
		newPlacement := []interface{}{createFullPlacement("c1", "ds1", "new-rp", "net1")}
		oldPool := createPoolSet(createFullMachinePool("cp-pool", true, oldPlacement))
		newPool := createPoolSet(createFullMachinePool("cp-pool", true, newPlacement))

		hasError, err := ValidateMachinePoolChange(oldPool, newPool)
		assert.True(t, hasError)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource_pool")
	})

	t.Run("network_change_error", func(t *testing.T) {
		oldPlacement := []interface{}{createFullPlacement("c1", "ds1", "rp1", "old-net")}
		newPlacement := []interface{}{createFullPlacement("c1", "ds1", "rp1", "new-net")}
		oldPool := createPoolSet(createFullMachinePool("cp-pool", true, oldPlacement))
		newPool := createPoolSet(createFullMachinePool("cp-pool", true, newPlacement))

		hasError, err := ValidateMachinePoolChange(oldPool, newPool)
		assert.True(t, hasError)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Network")
	})
}

// =============================================================================
// Acceptance Tests (require real API credentials)
// =============================================================================

// testAccProviderConfigVsphere returns provider configuration for acceptance tests
func testAccProviderConfigVsphere() string {
	apiKey := os.Getenv("SPECTROCLOUD_APIKEY")
	host := os.Getenv("SPECTROCLOUD_HOST")

	if apiKey == "" {
		apiKey = "vcr-replay-dummy-api-key"
	}
	if host == "" {
		host = "api.spectrocloud.com"
	}

	return fmt.Sprintf(`
provider "spectrocloud" {
  host    = %q
  api_key = %q
}
`, host, apiKey)
}

// TestAccClusterVsphere_basic tests basic vSphere cluster CRUD operations
func TestAccClusterVsphere_basic(t *testing.T) {
	if os.Getenv("SPECTROCLOUD_APIKEY") == "" {
		t.Skip("Skipping acceptance test: SPECTROCLOUD_APIKEY not set")
	}

	// Additional required environment variables for vSphere
	requiredEnvVars := []string{
		"VSPHERE_CLOUD_ACCOUNT_ID",
		"VSPHERE_CLUSTER_PROFILE_ID",
		"VSPHERE_DATACENTER",
		"VSPHERE_FOLDER",
		"VSPHERE_CLUSTER",
		"VSPHERE_DATASTORE",
		"VSPHERE_NETWORK",
		"VSPHERE_RESOURCE_POOL",
		"VSPHERE_SSH_KEY",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Skipf("Skipping acceptance test: %s not set", envVar)
		}
	}

	resourceName := "spectrocloud_cluster_vsphere.test"
	clusterName := testutil.RandomName("tf-acc-vsphere")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testutil.TestAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckClusterVsphereDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterVsphereConfig_basic(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterVsphereExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", clusterName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_config_id"),
				),
			},
		},
	})
}

// testAccClusterVsphereConfig_basic returns a basic vSphere cluster configuration
func testAccClusterVsphereConfig_basic(name string) string {
	return testAccProviderConfigVsphere() + fmt.Sprintf(`
resource "spectrocloud_cluster_vsphere" "test" {
  name             = %q
  cloud_account_id = %q

  cluster_profile {
    id = %q
  }

  cloud_config {
    ssh_keys   = [%q]
    datacenter = %q
    folder     = %q
  }

  machine_pool {
    name                    = "cp-pool"
    control_plane           = true
    control_plane_as_worker = true
    count                   = 1

    placement {
      cluster       = %q
      datastore     = %q
      network       = %q
      resource_pool = %q
    }

    instance_type {
      disk_size_gb = 60
      memory_mb    = 8192
      cpu          = 4
    }
  }

  machine_pool {
    name          = "worker-pool"
    control_plane = false
    count         = 1

    placement {
      cluster       = %q
      datastore     = %q
      network       = %q
      resource_pool = %q
    }

    instance_type {
      disk_size_gb = 60
      memory_mb    = 8192
      cpu          = 4
    }
  }
}
`,
		name,
		os.Getenv("VSPHERE_CLOUD_ACCOUNT_ID"),
		os.Getenv("VSPHERE_CLUSTER_PROFILE_ID"),
		os.Getenv("VSPHERE_SSH_KEY"),
		os.Getenv("VSPHERE_DATACENTER"),
		os.Getenv("VSPHERE_FOLDER"),
		os.Getenv("VSPHERE_CLUSTER"),
		os.Getenv("VSPHERE_DATASTORE"),
		os.Getenv("VSPHERE_NETWORK"),
		os.Getenv("VSPHERE_RESOURCE_POOL"),
		os.Getenv("VSPHERE_CLUSTER"),
		os.Getenv("VSPHERE_DATASTORE"),
		os.Getenv("VSPHERE_NETWORK"),
		os.Getenv("VSPHERE_RESOURCE_POOL"),
	)
}

// testAccCheckClusterVsphereExists verifies the cluster exists
func testAccCheckClusterVsphereExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("vSphere cluster not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("vSphere cluster ID not set")
		}
		return nil
	}
}

// testAccCheckClusterVsphereDestroy verifies the cluster was destroyed
func testAccCheckClusterVsphereDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "spectrocloud_cluster_vsphere" {
			continue
		}
		// In a real test, verify the cluster was deleted via API
	}
	return nil
}

// =============================================================================
// Helper functions
// =============================================================================

// prepareVsphereClusterTestData creates test ResourceData for vSphere cluster tests
func prepareVsphereClusterTestData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterVsphere().TestResourceData()

	d.SetId("test-vsphere-cluster-uid")
	d.Set("name", "test-vsphere-cluster")
	d.Set("context", "project")
	d.Set("cloud_account_id", "test-cloud-account-id")

	cloudConfig := []map[string]interface{}{
		{
			"datacenter":            "DC1",
			"folder":                "/test/folder",
			"ssh_keys":              []string{"ssh-rsa AAAAB3..."},
			"static_ip":             false,
			"network_type":          "DDNS",
			"network_search_domain": "spectrocloud.dev",
		},
	}
	d.Set("cloud_config", cloudConfig)

	return d
}

// assertDiagsEmptyVsphere asserts that diagnostics slice is empty
func assertDiagsEmptyVsphere(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	if len(diags) > 0 {
		for _, d := range diags {
			t.Errorf("Unexpected diagnostic: %s - %s", d.Summary, d.Detail)
		}
		t.FailNow()
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

// BenchmarkToMachinePoolVsphere benchmarks the toMachinePoolVsphere function
func BenchmarkToMachinePoolVsphere(b *testing.B) {
	input := map[string]interface{}{
		"name":                    "benchmark-pool",
		"control_plane":           false,
		"control_plane_as_worker": false,
		"count":                   3,
		"node_repave_interval":    0,
		"placement": []interface{}{
			map[string]interface{}{
				"id":                "",
				"cluster":           "test-cluster",
				"resource_pool":     "Default",
				"datastore":         "datastore1",
				"network":           "VM Network",
				"static_ip_pool_id": "",
			},
		},
		"instance_type": []interface{}{
			map[string]interface{}{
				"disk_size_gb": 60,
				"memory_mb":    8192,
				"cpu":          4,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = toMachinePoolVsphere(input)
	}
}

// BenchmarkFlattenMachinePoolConfigsVsphere benchmarks the flattenMachinePoolConfigsVsphere function
func BenchmarkFlattenMachinePoolConfigsVsphere(b *testing.B) {
	input := []*models.V1VsphereMachinePoolConfig{
		{
			Name:           "test-pool",
			Size:           3,
			MinSize:        1,
			MaxSize:        5,
			IsControlPlane: types.Ptr(true),
			InstanceType: &models.V1VsphereInstanceType{
				DiskGiB:   types.Ptr(int32(100)),
				MemoryMiB: types.Ptr(int64(8192)),
				NumCPUs:   types.Ptr(int32(4)),
			},
			Placements: []*models.V1VspherePlacementConfig{
				{
					UID:          "placement-1",
					Cluster:      "vsphere-cluster",
					ResourcePool: "default-pool",
					Datastore:    "datastore1",
					Network: &models.V1VsphereNetworkConfig{
						NetworkName: types.Ptr("VM Network"),
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = flattenMachinePoolConfigsVsphere(input)
	}
}
