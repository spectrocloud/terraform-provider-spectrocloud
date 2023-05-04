package schemas

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func PackSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uid": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"type": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "spectro",
					Description: "The type of the pack. The default value is `spectro`.",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the pack. The name must be unique within the cluster profile. ",
				},
				"registry_uid": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The registry UID of the pack. The registry UID is the unique identifier of the registry. ",
				},
				"tag": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The tag of the pack. The tag is the version of the pack.",
				},
				"values": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The values of the pack. The values are the configuration values of the pack. The values are specified in YAML format. ",
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
