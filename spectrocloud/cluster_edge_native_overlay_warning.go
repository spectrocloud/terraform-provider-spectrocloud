package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const overlayNetworkDisabledWarning = "Deploying in DHCP mode without overlay enabled can lead to IP changes that may cause cluster instability."

// appendOverlayNetworkDisabledWarning warns, on create only, when an Edge
// Native cluster's cloud_config.overlay_cidr_range is left empty (overlay
// networking disabled). It must not be called from Update: an existing
// cluster's overlay_cidr_range is ForceNew, so update is never the point
// where this condition is newly introduced.
func appendOverlayNetworkDisabledWarning(d *schema.ResourceData, diags *diag.Diagnostics) {
	cloudConfig, ok := d.Get("cloud_config").([]interface{})
	if !ok || len(cloudConfig) == 0 {
		return
	}
	cc, ok := cloudConfig[0].(map[string]interface{})
	if !ok {
		return
	}
	if cidr, ok := cc["overlay_cidr_range"].(string); !ok || cidr == "" {
		*diags = append(*diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Warning",
			Detail:   overlayNetworkDisabledWarning,
		})
	}
}
