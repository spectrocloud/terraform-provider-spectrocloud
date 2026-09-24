package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PLT-2436: per-node taint/label overrides for edge_host, so a single node
// pool can host non-uniform node roles (e.g. a witness/arbiter node kept
// non-schedulable while its peers stay schedulable) without splitting the
// pool. Wires straight through the SDK's V1EdgeNativeMachinePoolHostEntity
// (create/update) and V1EdgeNativeHost (read) Taints/AdditionalLabels
// fields, which already exist for exactly this purpose.

func TestToEdgeHosts_PerHostTaintsAndLabels(t *testing.T) {
	m := map[string]interface{}{
		"edge_host": schema.NewSet(resourceEdgeHostHash, []interface{}{
			map[string]interface{}{
				"host_name": "witness",
				"host_uid":  "witness-uid",
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "node.spectrocloud.com/witness",
						"value":  "true",
						"effect": "NoSchedule",
					},
				},
				"additional_labels": map[string]interface{}{
					"role": "witness",
				},
			},
			map[string]interface{}{
				"host_name": "primary",
				"host_uid":  "primary-uid",
			},
		}),
	}

	cloudConfig, err := toEdgeHosts(m)
	require.NoError(t, err)
	require.NotNil(t, cloudConfig)
	require.Len(t, cloudConfig.EdgeHosts, 2)

	var witness, primary *models.V1EdgeNativeMachinePoolHostEntity
	for _, h := range cloudConfig.EdgeHosts {
		switch *h.HostUID {
		case "witness-uid":
			witness = h
		case "primary-uid":
			primary = h
		}
	}
	require.NotNil(t, witness)
	require.NotNil(t, primary)

	require.Len(t, witness.Taints, 1)
	assert.Equal(t, "node.spectrocloud.com/witness", witness.Taints[0].Key)
	assert.Equal(t, "true", witness.Taints[0].Value)
	assert.Equal(t, "NoSchedule", witness.Taints[0].Effect)
	assert.Equal(t, map[string]string{"role": "witness"}, witness.AdditionalLabels)

	assert.Empty(t, primary.Taints)
	assert.Empty(t, primary.AdditionalLabels)
}

func TestFlattenEdgeNativePoolHost_TaintsAndLabels(t *testing.T) {
	hostUID := "witness-uid"
	host := &models.V1EdgeNativeHost{
		HostUID:  &hostUID,
		HostName: "witness",
		Taints: []*models.V1Taint{
			{Key: "node.spectrocloud.com/witness", Value: "true", Effect: "NoSchedule"},
		},
		AdditionalLabels: map[string]string{"role": "witness"},
	}

	rawHost := flattenEdgeNativePoolHost(host)
	require.NotNil(t, rawHost)

	taints, ok := rawHost["taints"].([]interface{})
	require.True(t, ok)
	require.Len(t, taints, 1)
	taint := taints[0].(map[string]interface{})
	assert.Equal(t, "node.spectrocloud.com/witness", taint["key"])
	assert.Equal(t, "true", taint["value"])
	assert.Equal(t, "NoSchedule", taint["effect"])

	assert.Equal(t, map[string]string{"role": "witness"}, rawHost["additional_labels"])
}

func TestFlattenEdgeNativePoolHost_NoTaintsOrLabelsOmitsKeys(t *testing.T) {
	hostUID := "plain-uid"
	host := &models.V1EdgeNativeHost{
		HostUID:  &hostUID,
		HostName: "plain",
	}

	rawHost := flattenEdgeNativePoolHost(host)
	require.NotNil(t, rawHost)
	_, hasTaints := rawHost["taints"]
	_, hasLabels := rawHost["additional_labels"]
	assert.False(t, hasTaints)
	assert.False(t, hasLabels)
}

func TestResourceEdgeHostHash_DetectsTaintAndLabelChanges(t *testing.T) {
	base := map[string]interface{}{
		"host_name": "witness",
		"host_uid":  "witness-uid",
	}
	withTaint := map[string]interface{}{
		"host_name": "witness",
		"host_uid":  "witness-uid",
		"taints": []interface{}{
			map[string]interface{}{
				"key":    "node.spectrocloud.com/witness",
				"value":  "true",
				"effect": "NoSchedule",
			},
		},
	}
	withLabel := map[string]interface{}{
		"host_name": "witness",
		"host_uid":  "witness-uid",
		"additional_labels": map[string]interface{}{
			"role": "witness",
		},
	}

	baseHash := resourceEdgeHostHash(base)
	taintHash := resourceEdgeHostHash(withTaint)
	labelHash := resourceEdgeHostHash(withLabel)

	assert.NotEqual(t, baseHash, taintHash, "adding a per-host taint must change the set element's hash so Terraform detects the diff")
	assert.NotEqual(t, baseHash, labelHash, "adding a per-host label must change the set element's hash so Terraform detects the diff")
	assert.NotEqual(t, taintHash, labelHash)
}
