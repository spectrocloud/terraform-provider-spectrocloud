package schemas

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func PackSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "For packs of type `spectro`, `helm`, and `manifest`, at least one pack must be specified.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uid": {
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    true,
					Description: "The unique identifier of the pack. The value can be looked up using the [`spectrocloud_pack`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs/data-sources/pack) data source. This value is required if the pack type is `spectro` and for `helm` if the chart is from a public helm registry.",
				},
				"type": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "spectro",
					Description: "The type of the pack. Allowed values are `spectro`, `manifest`, `helm`, or `oci`. The default value is spectro. If using an OCI registry for pack, set the type to `oci`.",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the pack. The name must be unique within the cluster profile. ",
				},
				"registry_uid": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The registry UID of the pack. The registry UID is the unique identifier of the registry. " +
						"This attribute is required if there is more than one registry that contains a pack with the same name. ",
				},
				"tag": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The tag of the pack. The tag is the version of the pack. This attribute is required if the pack type is `spectro` or `helm`. ",
				},
				"values": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The values of the pack. The values are the configuration values of the pack. The values are specified in YAML format. ",
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						// UI strips the trailing newline on save
						return strings.TrimSpace(old) == strings.TrimSpace(new)
					},
				},
				"manifest": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"uid": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of the manifest. The name must be unique within the pack. ",
							},
							"content": {
								Type:     schema.TypeString,
								Required: true,
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									// UI strips the trailing newline on save
									return strings.TrimSpace(old) == strings.TrimSpace(new)
								},
								Description: "The content of the manifest. The content is the YAML content of the manifest. ",
							},
						},
					},
				},
			},
		},
	}
}
