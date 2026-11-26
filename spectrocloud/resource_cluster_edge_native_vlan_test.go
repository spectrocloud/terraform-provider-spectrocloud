package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// PLT-2511: edge host NIC VLAN (vlanId, secondaryIps).

func vlanEdgeHostConfig() map[string]interface{} {
	return map[string]interface{}{
		"host_uid":      "vlan-host-uid",
		"nic_name":      "ens192",
		"static_ip":     "10.3.12.164",
		"subnet_mask":   "255.255.255.224",
		"vlan_id":       200,
		"secondary_ips": schema.NewSet(schema.HashString, []interface{}{"10.3.12.166/27", "10.3.12.165/27"}),
	}
}

func TestToEdgeHostsSendsVlan(t *testing.T) {
	m := map[string]interface{}{
		"edge_host": schema.NewSet(resourceEdgeHostHash, []interface{}{
			vlanEdgeHostConfig(),
			map[string]interface{}{"host_uid": "plain-host-uid", "nic_name": "ens160"},
		}),
	}

	cloudConfig, err := toEdgeHosts(m)
	require.NoError(t, err)
	require.Len(t, cloudConfig.EdgeHosts, 2)

	hosts := map[string]*models.V1EdgeNativeMachinePoolHostEntity{}
	for _, h := range cloudConfig.EdgeHosts {
		hosts[*h.HostUID] = h
	}

	vlan := hosts["vlan-host-uid"].Nic
	require.NotNil(t, vlan.VlanID)
	assert.Equal(t, int32(200), *vlan.VlanID)
	assert.ElementsMatch(t, []string{"10.3.12.165/27", "10.3.12.166/27"}, vlan.SecondaryIps)
	assert.Equal(t, "ens192", vlan.NicName)
	assert.Equal(t, "10.3.12.164", vlan.IP)
	assert.Equal(t, "255.255.255.224", vlan.Subnet)

	plain := hosts["plain-host-uid"].Nic
	assert.Nil(t, plain.VlanID, "a host without vlan_id sends no vlanId")
	assert.Empty(t, plain.SecondaryIps)
}

func TestFlattenEdgeNativePoolHostReadsVlan(t *testing.T) {
	withVlan := flattenEdgeNativePoolHost(&models.V1EdgeNativeHost{
		HostUID: types.Ptr("vlan-host-uid"),
		Nic: &models.V1Nic{
			NicName:      "ens192",
			IP:           "10.3.12.164",
			Subnet:       "255.255.255.224",
			VlanID:       types.Ptr(int32(200)),
			SecondaryIps: []string{"10.3.12.165/27"},
		},
	})
	assert.Equal(t, 200, withVlan["vlan_id"])
	assert.Equal(t, []string{"10.3.12.165/27"}, withVlan["secondary_ips"])

	noVlan := flattenEdgeNativePoolHost(&models.V1EdgeNativeHost{
		HostUID: types.Ptr("plain-host-uid"),
		Nic:     &models.V1Nic{NicName: "ens160"},
	})
	assert.NotContains(t, noVlan, "vlan_id", "a host without a VLAN reads back as before")
	assert.NotContains(t, noVlan, "secondary_ips")
}

func TestResourceEdgeHostHashVlan(t *testing.T) {
	config := vlanEdgeHostConfig()

	// What Read produces for the same host must hash the same as the configuration, or every plan shows a diff.
	readBack := flattenEdgeNativePoolHost(&models.V1EdgeNativeHost{
		HostUID: types.Ptr("vlan-host-uid"),
		Nic: &models.V1Nic{
			NicName:      "ens192",
			IP:           "10.3.12.164",
			Subnet:       "255.255.255.224",
			VlanID:       types.Ptr(int32(200)),
			SecondaryIps: []string{"10.3.12.165/27", "10.3.12.166/27"},
		},
	})
	assert.Equal(t, resourceEdgeHostHash(config), resourceEdgeHostHash(readBack), "read-back must match the configuration")

	changedVlan := vlanEdgeHostConfig()
	changedVlan["vlan_id"] = 300
	assert.NotEqual(t, resourceEdgeHostHash(config), resourceEdgeHostHash(changedVlan), "a vlan_id change must be detected")

	changedIPs := vlanEdgeHostConfig()
	changedIPs["secondary_ips"] = schema.NewSet(schema.HashString, []interface{}{"10.3.12.165/27"})
	assert.NotEqual(t, resourceEdgeHostHash(config), resourceEdgeHostHash(changedIPs), "a secondary_ips change must be detected")

	// A host without VLAN fields hashes as before (no vlan_id or secondary_ips contribution).
	plain := map[string]interface{}{"host_uid": "plain-host-uid", "nic_name": "ens160"}
	plainZero := map[string]interface{}{"host_uid": "plain-host-uid", "nic_name": "ens160", "vlan_id": 0, "secondary_ips": []string{}}
	assert.Equal(t, resourceEdgeHostHash(plain), resourceEdgeHostHash(plainZero))
}
