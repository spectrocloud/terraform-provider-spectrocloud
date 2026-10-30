package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAppendOverlayNetworkDisabledWarning_OverlayDisabled is the PLT-2482
// direct guard on the helper: an empty overlay_cidr_range must append
// exactly one Warning diagnostic with the ticket-specified text.
func TestAppendOverlayNetworkDisabledWarning_OverlayDisabled(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"vip":                "10.0.0.100",
			"overlay_cidr_range": "",
		},
	}))

	var diags diag.Diagnostics
	appendOverlayNetworkDisabledWarning(d, &diags)

	require.Len(t, diags, 1)
	assert.Equal(t, diag.Warning, diags[0].Severity)
	assert.Equal(t, overlayNetworkDisabledWarning, diags[0].Detail)
}

// TestAppendOverlayNetworkDisabledWarning_OverlayEnabled is the symmetric
// guard: a non-empty overlay_cidr_range must append no diagnostic at all.
func TestAppendOverlayNetworkDisabledWarning_OverlayEnabled(t *testing.T) {
	d := resourceClusterEdgeNative().TestResourceData()
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"vip":                "10.0.0.100",
			"overlay_cidr_range": "100.64.192.0/24",
		},
	}))

	var diags diag.Diagnostics
	appendOverlayNetworkDisabledWarning(d, &diags)

	assert.Empty(t, diags)
}
