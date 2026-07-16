package virtualmachineinstance

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// networks_volumes_test.go — Batch 11.
// Covers 9 helpers across networks.go and volumes.go: expand/flatten for
// both, plus the container_disk / cloud_init / data_volume switch inside
// expandVolumeSourceForHapi.

// ---------------------------------------------------------------------------
// Networks
// ---------------------------------------------------------------------------

func TestExpandNetworksToVM(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got := expandNetworksToVM(nil)
		assert.Empty(t, got)
	})

	t.Run("pod network default", func(t *testing.T) {
		got := expandNetworksToVM([]interface{}{
			map[string]interface{}{
				"name": "default",
				"network_source": []interface{}{
					map[string]interface{}{
						"pod": []interface{}{
							map[string]interface{}{
								"vm_network_cidr":      "10.244.0.0/24",
								"vm_ipv6_network_cidr": "fd00::/64",
							},
						},
					},
				},
			},
		})
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Pod)
		assert.Equal(t, "10.244.0.0/24", got[0].Pod.VMNetworkCIDR)
	})

	t.Run("pod network with empty inner list uses zero-value", func(t *testing.T) {
		got := expandNetworksToVM([]interface{}{
			map[string]interface{}{
				"name": "default",
				"network_source": []interface{}{
					map[string]interface{}{
						"pod": []interface{}{}, // len=0 branch
					},
				},
			},
		})
		require.Len(t, got, 1)
		// Empty pod list attaches an empty V1VMPodNetwork.
		assert.NotNil(t, got[0].Pod)
	})

	t.Run("multus wins over pod", func(t *testing.T) {
		got := expandNetworksToVM([]interface{}{
			map[string]interface{}{
				"name": "secondary",
				"network_source": []interface{}{
					map[string]interface{}{
						"pod":    []interface{}{map[string]interface{}{"vm_network_cidr": "10.0.0.0/24"}},
						"multus": []interface{}{map[string]interface{}{"network_name": "net-attach", "default": true}},
					},
				},
			},
		})
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Multus)
		assert.Nil(t, got[0].Pod)
	})
}

func TestExpandPodNetworkToVM(t *testing.T) {
	// Empty input → empty struct.
	got := expandPodNetworkToVM(nil)
	require.NotNil(t, got)

	got = expandPodNetworkToVM([]interface{}{
		map[string]interface{}{"vm_network_cidr": "10.244.0.0/24", "vm_ipv6_network_cidr": "fd00::/64"},
	})
	require.NotNil(t, got)
	assert.Equal(t, "10.244.0.0/24", got.VMNetworkCIDR)
}

func TestExpandMultusNetworkToVM(t *testing.T) {
	assert.Nil(t, expandMultusNetworkToVM(nil))

	got := expandMultusNetworkToVM([]interface{}{
		map[string]interface{}{"network_name": "n1", "default": true},
	})
	require.NotNil(t, got)
	require.NotNil(t, got.NetworkName)
	assert.Equal(t, "n1", *got.NetworkName)
	assert.True(t, got.Default)
}

func TestFlattenNetworksFromVM(t *testing.T) {
	assert.Nil(t, flattenNetworksFromVM(nil))

	name := "default"
	got := flattenNetworksFromVM([]*models.V1VMNetwork{
		nil,
		{Name: &name, Pod: &models.V1VMPodNetwork{VMNetworkCIDR: "10.0.0.0/24"}},
	})
	require.Len(t, got, 2)
	m := got[1].(map[string]interface{})
	assert.Equal(t, "default", m["name"])
	assert.Contains(t, m, "network_source")
}

func TestFlattenNetworkSourceFromVM(t *testing.T) {
	// nil input still returns a slice with an empty map.
	got := flattenNetworkSourceFromVM(nil)
	require.Len(t, got, 1)

	// Both pod and multus set → both keys present.
	nName := "n1"
	got = flattenNetworkSourceFromVM(&models.V1VMNetwork{
		Pod:    &models.V1VMPodNetwork{VMNetworkCIDR: "10.0.0.0/24", VMIPV6NetworkCIDR: "fd00::/64"},
		Multus: &models.V1VMMultusNetwork{NetworkName: &nName, Default: true},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Contains(t, m, "pod")
	assert.Contains(t, m, "multus")
}

// ---------------------------------------------------------------------------
// Volumes
// ---------------------------------------------------------------------------

func TestExpandVolumes(t *testing.T) {
	assert.Empty(t, expandVolumes(nil))

	got := expandVolumes([]interface{}{
		map[string]interface{}{
			"name": "root",
			"volume_source": []interface{}{
				map[string]interface{}{
					// container_disk supports []interface{} (see expandVolumeSourceForHapi)
					// as one of its accepted shapes.
					"container_disk": []interface{}{
						map[string]interface{}{"image_url": "docker.io/fedora:latest"},
					},
				},
			},
		},
	})
	require.Len(t, got, 1)
	require.NotNil(t, got[0].Name)
	assert.Equal(t, "root", *got[0].Name)
}

func TestExpandVolumeSourceForHapi(t *testing.T) {
	t.Run("nil vol short-circuits", func(t *testing.T) {
		expandVolumeSourceForHapi(nil, nil) // must not panic
	})

	t.Run("container_disk via []interface{}", func(t *testing.T) {
		vol := &models.V1VMVolume{}
		expandVolumeSourceForHapi(map[string]interface{}{
			"container_disk": []interface{}{
				map[string]interface{}{"image_url": "docker.io/fedora:latest"},
			},
		}, vol)
		require.NotNil(t, vol.ContainerDisk)
	})

	t.Run("data_volume", func(t *testing.T) {
		vol := &models.V1VMVolume{}
		expandVolumeSourceForHapi(map[string]interface{}{
			"data_volume": []interface{}{map[string]interface{}{"name": "bootvol"}},
		}, vol)
		require.NotNil(t, vol.DataVolume)
		require.NotNil(t, vol.DataVolume.Name)
		assert.Equal(t, "bootvol", *vol.DataVolume.Name)
	})

	t.Run("cloud_init_config_drive", func(t *testing.T) {
		vol := &models.V1VMVolume{}
		expandVolumeSourceForHapi(map[string]interface{}{
			"cloud_init_config_drive": []interface{}{map[string]interface{}{
				"user_data":           "u",
				"user_data_base64":    "ub",
				"network_data":        "n",
				"network_data_base64": "nb",
			}},
		}, vol)
		require.NotNil(t, vol.CloudInitConfigDrive)
		assert.Equal(t, "u", vol.CloudInitConfigDrive.UserData)
	})

	t.Run("cloud_init_no_cloud via schema.Set", func(t *testing.T) {
		vol := &models.V1VMVolume{}
		expandVolumeSourceForHapi(map[string]interface{}{
			"cloud_init_no_cloud": []interface{}{map[string]interface{}{
				"user_data":        "u",
				"user_data_base64": "ub",
			}},
		}, vol)
		require.NotNil(t, vol.CloudInitNoCloud)
	})
}

func TestFlattenVolumesFromVM(t *testing.T) {
	assert.Nil(t, flattenVolumesFromVM(nil))

	name := "root"
	image := "docker.io/fedora:latest"
	got := flattenVolumesFromVM([]*models.V1VMVolume{
		nil,
		{Name: &name, ContainerDisk: &models.V1VMContainerDiskSource{Image: &image}},
	})
	require.Len(t, got, 2)
	m := got[1].(map[string]interface{})
	assert.Equal(t, "root", m["name"])
	assert.Contains(t, m, "volume_source")
}

func TestFlattenVolumeSourceFromVM(t *testing.T) {
	// nil returns a slice with an empty map.
	got := flattenVolumeSourceFromVM(nil)
	require.Len(t, got, 1)

	dvName := "bootvol"
	image := "docker.io/fedora:latest"
	got = flattenVolumeSourceFromVM(&models.V1VMVolume{
		DataVolume:           &models.V1VMCoreDataVolumeSource{Name: &dvName},
		ContainerDisk:        &models.V1VMContainerDiskSource{Image: &image},
		CloudInitNoCloud:     &models.V1VMCloudInitNoCloudSource{UserData: "u", UserDataBase64: "ub"},
		CloudInitConfigDrive: &models.V1VMCloudInitConfigDriveSource{UserData: "u"},
		Ephemeral:            &models.V1VMEphemeralVolumeSource{},
		// EmptyDisk / HostDisk are constructed with the right pointer
		// types below.
		EmptyDisk: func() *models.V1VMEmptyDiskSource {
			q := models.V1VMQuantity("1Gi")
			return &models.V1VMEmptyDiskSource{Capacity: &q}
		}(),
		HostDisk: func() *models.V1VMHostDisk {
			p, ty := "/tmp", "Disk"
			return &models.V1VMHostDisk{Path: &p, Type: &ty}
		}(),
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	for _, k := range []string{"data_volume", "container_disk", "cloud_init_no_cloud",
		"cloud_init_config_drive", "ephemeral", "empty_disk", "host_disk"} {
		assert.Contains(t, m, k)
	}
}
