package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func TestFlattenMachinePoolConfigsEdgeVsphere(t *testing.T) {
	tests := []struct {
		name     string
		input    []*models.V1VsphereMachinePoolConfig
		expected []interface{}
	}{
		{
			name:     "Nil machinePools input",
			input:    nil,
			expected: make([]interface{}, 0),
		},
		{
			name:     "Empty machinePools input",
			input:    []*models.V1VsphereMachinePoolConfig{},
			expected: make([]interface{}, 0),
		},
		{
			name: "Single machine pool with all fields populated",
			input: []*models.V1VsphereMachinePoolConfig{
				{
					AdditionalLabels: map[string]string{"env": "prod"},
					Taints: []*models.V1Taint{
						{
							Effect:    "NoSchedule",
							Key:       "key",
							TimeAdded: models.V1Time{},
							Value:     "np",
						},
					},
					IsControlPlane: ptr.To((true),
						NodeRepaveInterval:      30,
					UseControlPlaneAsWorker: true,
					Name:                    "pool1",
					Size:                    3,
					InstanceType: &models.V1VsphereInstanceType{
					DiskGiB:   int32Ptr(100),
					MemoryMiB: int64Ptr(8192),
					NumCPUs:   int32Ptr(4),
				},
					Placements: []*models.V1VspherePlacementConfig{
				{
					UID:          "placement1",
					Cluster:      "cluster1",
					ResourcePool: "resourcepool1",
					Datastore:    "datastore1",
					Network: &models.V1VsphereNetworkConfig{
					NetworkName:   ptr.To("network1"),
					ParentPoolRef: &models.V1ObjectReference{UID: "pool1"},
				},
				},
				},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"control_plane_as_worker": true,
					"name":                    "pool1",
					"count":                   3,
					"instance_type": []interface{}{
						map[string]interface{}{
							"disk_size_gb": 100,
							"memory_mb":    8192,
							"cpu":          4,
						},
					},
					"placement": []interface{}{
						map[string]interface{}{
							"id":                "placement1",
							"cluster":           "cluster1",
							"resource_pool":     "resourcepool1",
							"datastore":         "datastore1",
							"network":           "network1",
							"static_ip_pool_id": "pool1",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = flattenMachinePoolConfigsEdgeVsphere(tt.input)
		})
	}
}

func TestToMachinePoolEdgeVsphere(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]interface{}
		expected    *models.V1VsphereMachinePoolConfigEntity
		expectError bool
	}{
		{
			name: "Valid input with worker nodes",
			input: map[string]interface{}{
				"control_plane":           false,
				"control_plane_as_worker": false,
				"name":                    "worker-pool",
				"count":                   3,
				"instance_type": []interface{}{
					map[string]interface{}{
						"disk_size_gb": 100,
						"memory_mb":    8192,
						"cpu":          4,
					},
				},
				"placement": []interface{}{
					map[string]interface{}{
						"id":                "placement1",
						"cluster":           "cluster1",
						"resource_pool":     "resourcepool1",
						"datastore":         "datastore1",
						"network":           "network1",
						"static_ip_pool_id": "pool1",
					},
				},
				"node_repave_interval": 24,
			},
			expected: &models.V1VsphereMachinePoolConfigEntity{
				CloudConfig: &models.V1VsphereMachinePoolCloudConfigEntity{
					Placements: []*models.V1VspherePlacementConfigEntity{
						{
							UID:          "placement1",
							Cluster:      "cluster1",
							ResourcePool: "resourcepool1",
							Datastore:    "datastore1",
							Network: &models.V1VsphereNetworkConfigEntity{
								NetworkName:   ptr.To("network1"),
								ParentPoolUID: "pool1",
								StaticIP:      true,
							},
						},
					},
					InstanceType: &models.V1VsphereInstanceType{
						DiskGiB:   ptr.To(int32(100)),
						MemoryMiB: ptr.To(int64(8192)),
						NumCPUs:   ptr.To(int32(4)),
					},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					IsControlPlane:     false,
					Labels:             []string{"worker"},
					Name:               ptr.To("worker-pool"),
					Size:               ptr.To(int32(3)),
					NodeRepaveInterval: 24,
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "",
					},
					UseControlPlaneAsWorker: false,
				},
			},
			expectError: false,
		},
		{
			name: "Valid input with control plane nodes",
			input: map[string]interface{}{
				"control_plane":           true,
				"control_plane_as_worker": true,
				"name":                    "control-plane-pool",
				"count":                   1,
				"instance_type": []interface{}{
					map[string]interface{}{
						"disk_size_gb": 200,
						"memory_mb":    16384,
						"cpu":          8,
					},
				},
				"placement": []interface{}{
					map[string]interface{}{
						"id":                "placement2",
						"cluster":           "cluster2",
						"resource_pool":     "resourcepool2",
						"datastore":         "datastore2",
						"network":           "network2",
						"static_ip_pool_id": "",
					},
				},
				"node_repave_interval": 12,
			},
			expected: &models.V1VsphereMachinePoolConfigEntity{
				CloudConfig: &models.V1VsphereMachinePoolCloudConfigEntity{
					Placements: []*models.V1VspherePlacementConfigEntity{
						{
							UID:          "placement2",
							Cluster:      "cluster2",
							ResourcePool: "resourcepool2",
							Datastore:    "datastore2",
							Network: &models.V1VsphereNetworkConfigEntity{
								NetworkName:   ptr.To("network2"),
								ParentPoolUID: "",
								StaticIP:      false,
							},
						},
					},
					InstanceType: &models.V1VsphereInstanceType{
						DiskGiB:   ptr.To(int32(200)),
						MemoryMiB: ptr.To(int64(16384)),
						NumCPUs:   ptr.To(int32(8)),
					},
				},
				PoolConfig: &models.V1MachinePoolConfigEntity{
					IsControlPlane: true,
					Labels:         []string{"control-plane"},
					Name:           ptr.To("control-plane-pool"),
					Size:           ptr.To(int32(1)),
					UpdateStrategy: &models.V1UpdateStrategy{
						Type: "",
					},
					UseControlPlaneAsWorker: true,
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _ = toMachinePoolEdgeVsphere(tt.input)

		})
	}
}
