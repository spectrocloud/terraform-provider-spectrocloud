package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// OverrideClusterAPIConfigSchema returns the schema for YAML CAPI overrides at cloud config (cluster) level.
func OverrideClusterAPIConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "YAML override for CAPI properties at cluster level. Overrides pack-level and Palette-managed values.",
	}
}

// OverrideClusterAPIConfigMachinePoolSchema returns the schema for YAML CAPI overrides at machine pool level.
func OverrideClusterAPIConfigMachinePoolSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "YAML override for CAPI properties at machine pool level. Overrides pack-level and Palette-managed values.",
	}
}
