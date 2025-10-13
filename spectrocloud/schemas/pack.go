package schemas

import (
	"fmt"
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
					Description: "The unique identifier of the pack. The value can be looked up using the [`spectrocloud_pack`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs/data-sources/pack) data source. This value is required if the pack type is `spectro` and for `helm` if the chart is from a public helm registry. If not provided, all of `name`, `tag`, and `registry_uid` must be specified to resolve the pack UID internally.",
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
						"This attribute is required if there is more than one registry that contains a pack with the same name. " +
						"If `uid` is not provided, this field is required along with `name` and `tag` to resolve the pack UID internally. " +
						"Either `registry_uid` or `registry_name` can be specified, but not both.",
				},
				"registry_name": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The registry name of the pack. The registry name is the human-readable name of the registry. " +
						"This attribute can be used instead of `registry_uid` for better readability. " +
						"If `uid` is not provided, this field can be used along with `name` and `tag` to resolve the pack UID internally. " +
						"Either `registry_uid` or `registry_name` can be specified, but not both.",
				},
				"tag": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The tag of the pack. The tag is the version of the pack. This attribute is required if the pack type is `spectro` or `helm`. " +
						"If `uid` is not provided, this field is required along with `name` and `registry_uid` (or `registry_name`) to resolve the pack UID internally.",
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

// ValidatePackUIDOrResolutionFields validates that either uid is provided
// OR all of name, tag, and registry_uid/registry_name are specified for pack resolution.
func ValidatePackUIDOrResolutionFields(packData map[string]interface{}) error {
	uid := ""
	if packData["uid"] != nil {
		uid = packData["uid"].(string)
	}

	name := ""
	if packData["name"] != nil {
		name = packData["name"].(string)
	}

	tag := ""
	if packData["tag"] != nil {
		tag = packData["tag"].(string)
	}

	registryUID := ""
	if packData["registry_uid"] != nil {
		registryUID = packData["registry_uid"].(string)
	}

	registryName := ""
	if packData["registry_name"] != nil {
		registryName = packData["registry_name"].(string)
	}

	packType := ""
	if packData["type"] != nil {
		packType = packData["type"].(string)
	}

	// Validate that both registry_uid and registry_name are not provided together
	if registryUID != "" && registryName != "" {
		return fmt.Errorf("pack %s: only one of 'registry_uid' or 'registry_name' can be specified, not both", name)
	}

	// Skip validation for manifest packs as they have special handling
	if packType == "manifest" {
		return nil
	}

	// If uid is provided, validation passes
	if uid != "" {
		return nil
	}

	// If uid is not provided, check if all required fields for resolution are present
	missingFields := make([]string, 0)

	if name == "" {
		missingFields = append(missingFields, "name")
	}
	if tag == "" {
		missingFields = append(missingFields, "tag")
	}
	if registryUID == "" && registryName == "" {
		missingFields = append(missingFields, "registry_uid or registry_name")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("pack %s: either 'uid' must be provided, or all of the following fields must be specified for pack resolution: %s. Missing: %s",
			name, "name, tag, registry_uid (or registry_name)", strings.Join(missingFields, ", "))
	}

	return nil
}
