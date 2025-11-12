package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ClusterTemplateSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "[Tech Preview] The ID of the cluster template. When a cluster is launched using a template, the packs configuration is automatically derived from the template. Cluster template does not support day 2 operations - changing the template after cluster creation is not allowed.",
	}
}
