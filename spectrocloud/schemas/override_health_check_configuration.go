package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// OverrideHealthCheckConfigurationSchema returns the schema for YAML Machine Health Check overrides at machine pool level.
func OverrideHealthCheckConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: "YAML override for Machine Health Check configuration at the node pool level (control plane and worker pools). " +
			"Accepts CAPI MachineHealthCheck fields such as maxUnhealthy, nodeStartupTimeout, and unhealthyConditions. " +
			"Falls back to Palette defaults when unset. Still respects the project/tenant Cluster Auto Remediation setting. " +
			"Changing this value may repave your nodes.",
	}
}
