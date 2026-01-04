package schemas

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func AppPackSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "A list of packs to be applied to the application profile.",
		Set:         resourceAppPackHash,
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

func resourceAppPackHash(v interface{}) int {
	m := v.(map[string]interface{})
	var buf bytes.Buffer

	// Primary identifier - name is required
	if val, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("name-%s-", val.(string)))
	}

	// if val, ok := m["uid"]; ok && val != nil && val != "" {
	// 	buf.WriteString(fmt.Sprintf("uid-%s-", val.(string)))
	// }

	// Pack type (optional, default "spectro")
	if val, ok := m["type"]; ok && val != nil && val != "" {
		buf.WriteString(fmt.Sprintf("type-%s-", val.(string)))
	} else {
		buf.WriteString("type-spectro-") // Default value
	}

	// Tag/version identifier
	if val, ok := m["tag"]; ok && val != nil && val != "" {
		buf.WriteString(fmt.Sprintf("tag-%s-", val.(string)))
	}

	// Registry identifier - use registry_uid if available, otherwise registry_name
	if val, ok := m["registry_uid"]; ok && val != nil && val != "" {
		buf.WriteString(fmt.Sprintf("registry_uid-%s-", val.(string)))
	} else if val, ok := m["registry_name"]; ok && val != nil && val != "" {
		buf.WriteString(fmt.Sprintf("registry_name-%s-", val.(string)))
	}

	// Source app tier
	if val, ok := m["source_app_tier"]; ok && val != nil && val != "" {
		buf.WriteString(fmt.Sprintf("source_app_tier-%s-", val.(string)))
	}

	// Install order (optional, default 0)
	if val, ok := m["install_order"]; ok {
		if intVal, ok := val.(int); ok {
			buf.WriteString(fmt.Sprintf("install_order-%d-", intVal))
		}
	} else {
		buf.WriteString("install_order-0-") // Default value
	}

	// Properties map
	// Properties map - FIX: removed duplicate dead code
	if val, ok := m["properties"]; ok && val != nil {
		if props, ok := val.(map[string]interface{}); ok && len(props) > 0 {
			// Sort keys for deterministic hashing
			keys := make([]string, 0, len(props))
			for k := range props {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if v, ok := props[k].(string); ok {
					buf.WriteString(fmt.Sprintf("properties-%s-%s-", k, v))
				}
			}
		}
	}

	// Values - normalize by trimming whitespace (matching DiffSuppressFunc behavior)
	if val, ok := m["values"]; ok && val != nil {
		valuesStr := ""
		if v, ok := val.(string); ok {
			valuesStr = strings.TrimSpace(v)
		}
		if valuesStr != "" {
			buf.WriteString(fmt.Sprintf("values-%s-", valuesStr))
		}
	}

	// Manifest list - include in hash since it affects pack identity
	// Since manifest is TypeList (order matters), we preserve order in hash
	if val, ok := m["manifest"]; ok && val != nil {
		if manifestList, ok := val.([]interface{}); ok && len(manifestList) > 0 {
			for i, manifest := range manifestList {
				if m, ok := manifest.(map[string]interface{}); ok {
					manifestName := ""
					manifestContent := ""
					if name, ok := m["name"].(string); ok {
						manifestName = name
					}
					if content, ok := m["content"].(string); ok {
						// Normalize content by trimming whitespace (matching DiffSuppressFunc)
						manifestContent = strings.TrimSpace(content)
					}
					// Include index to preserve order
					buf.WriteString(fmt.Sprintf("manifest-%d-name-%s-content-%s-", i, manifestName, manifestContent))
				}
			}
		}
	}

	// DO NOT include "uid" - it's a computed field and should not affect hash

	return int(hash(buf.String()))
}
