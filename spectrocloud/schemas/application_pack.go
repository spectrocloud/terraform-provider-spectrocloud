package schemas

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func AppPackSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		Description: "A list of packs to be applied to the application profile.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The type of Pack. Allowed values are `container`, `helm`, `manifest`, or `operator-instance`.",
					Default:     "spectro",
				},
				"source_app_tier": {
					Type:        schema.TypeString,
					Description: "The unique id of the pack to be used as the source for the pack.",
					Optional:    true,
				},
				"registry_uid": {
					Type:        schema.TypeString,
					Description: "The unique id of the registry to be used for the pack. Either `registry_uid` or `registry_name` can be specified, but not both.",
					Optional:    true,
				},
				"registry_name": {
					Type:        schema.TypeString,
					Description: "The name of the registry to be used for the pack. This can be used instead of `registry_uid` for better readability. Either `registry_uid` or `registry_name` can be specified, but not both.",
					Optional:    true,
				},
				"uid": {
					Type:        schema.TypeString,
					Description: "The unique id of the pack. This is a computed field and is not required to be set.",
					Computed:    true,
					Optional:    true,
				},
				"name": {
					Type:        schema.TypeString,
					Description: "The name of the specified pack.",
					Required:    true,
				},
				"properties": {
					Type:        schema.TypeMap,
					Optional:    true,
					Description: "The various properties required by different database tiers eg: `databaseName` and `databaseVolumeSize` size for Redis etc.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"install_order": {
					Type:        schema.TypeInt,
					Description: "The installation priority order of the app profile. The order of priority goes from lowest number to highest number. For example, a value of `-3` would be installed before an app profile with a higher number value. No upper and lower limits exist, and you may specify positive and negative integers. The default value is `0`. ",
					Default:     0,
					Optional:    true,
				},
				"manifest": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The manifest of the pack.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"uid": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"name": {
								Type:        schema.TypeString,
								Description: "The name of the manifest.",
								Required:    true,
							},
							"content": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The content of the manifest.",
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									// UI strips the trailing newline on save
									return strings.TrimSpace(old) == strings.TrimSpace(new)
								},
							},
						},
					},
				},
				"tag": {
					Type:        schema.TypeString,
					Description: "The identifier or version to label the pack.",
					Optional:    true,
				},
				"values": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The values to be used for the pack. This is a stringified JSON object.",
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						// UI strips the trailing newline on save
						return strings.TrimSpace(old) == strings.TrimSpace(new)
					},
				},
			},
		},
	}
}
