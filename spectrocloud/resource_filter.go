package spectrocloud

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceFilter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFilterCreate,
		ReadContext:   resourceFilterRead,
		UpdateContext: resourceFilterUpdate,
		DeleteContext: resourceFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFilterImport,
		},
		Description: "A resource for creating and managing filters.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 3,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceFilterResourceV2().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceFilterStateUpgradeV2,
				Version: 2,
			},
		},
		Schema: map[string]*schema.Schema{
			"metadata": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Metadata of the filter.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the filter.",
						},
					},
				},
			},
			"spec": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Specification of the filter.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_group": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Filter group of the filter.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"conjunction": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"and", "or"}, false),
										Description:  "Conjunction operation of the filter group. Valid values are 'and' and 'or'.",
									},
									"filters": {
										Type:        schema.TypeSet,
										Required:    true,
										Set:         resourceFilterItemHash,
										Description: "List of filters in the filter group.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Key of the filter.",
												},
												"negation": {
													Type:        schema.TypeBool,
													Optional:    true,
													Default:     false,
													Description: "Negation flag of the filter condition.",
												},
												"operator": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"eq"}, false),
													Description:  "Operator of the filter. Valid values are 'eq'.",
												},
												"values": {
													Type:        schema.TypeList,
													Required:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Description: "Values of the filter.",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceFilterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	metadata := d.Get("metadata").([]interface{})
	spec := d.Get("spec").([]interface{})

	tagFilter := &models.V1TagFilter{
		Metadata: expandMetadata(metadata),
		Spec:     expandSpec(spec),
	}

	uid, err := c.CreateTagFilter(tagFilter)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*uid.UID)
	return resourceFilterRead(ctx, d, m)
}

func resourceFilterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	uid := d.Id()

	tagFilterSummary, err := c.GetTagFilter(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	if err := d.Set("metadata", flattenMetadata(tagFilterSummary.Metadata)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("spec", flattenSpec(tagFilterSummary.Spec)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFilterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	tagFilter := &models.V1TagFilter{
		Metadata: expandMetadata(d.Get("metadata").([]interface{})),
		Spec:     expandSpec(d.Get("spec").([]interface{})),
	}

	err := c.UpdateTagFilter(d.Id(), tagFilter)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFilterRead(ctx, d, m)
}

func resourceFilterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	err := c.DeleteTagFilter(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

// resourceFilterResourceV2 returns the schema for version 2 of the resource
// where filters was TypeList instead of TypeSet
func resourceFilterResourceV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metadata": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Metadata of the filter.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the filter.",
						},
					},
				},
			},
			"spec": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Specification of the filter.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_group": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Filter group of the filter.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"conjunction": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"and", "or"}, false),
										Description:  "Conjunction operation of the filter group. Valid values are 'and' and 'or'.",
									},
									"filters": {
										Type:        schema.TypeList, // V2: TypeList
										Required:    true,
										Description: "List of filters in the filter group.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Key of the filter.",
												},
												"negation": {
													Type:        schema.TypeBool,
													Optional:    true,
													Default:     false,
													Description: "Negation flag of the filter condition.",
												},
												"operator": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"eq"}, false),
													Description:  "Operator of the filter. Valid values are 'eq'.",
												},
												"values": {
													Type:        schema.TypeList,
													Required:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Description: "Values of the filter.",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// resourceFilterStateUpgradeV2 migrates state from version 2 to version 3
// Converts filters from TypeList to TypeSet
func resourceFilterStateUpgradeV2(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Upgrading filter state from version 2 to 3")

	// Navigate to spec -> filter_group -> filters
	if specRaw, exists := rawState["spec"]; exists {
		if specList, ok := specRaw.([]interface{}); ok && len(specList) > 0 {
			if specMap, ok := specList[0].(map[string]interface{}); ok {
				if filterGroupRaw, exists := specMap["filter_group"]; exists {
					if filterGroupList, ok := filterGroupRaw.([]interface{}); ok && len(filterGroupList) > 0 {
						if filterGroupMap, ok := filterGroupList[0].(map[string]interface{}); ok {
							if filtersRaw, exists := filterGroupMap["filters"]; exists {
								if filtersList, ok := filtersRaw.([]interface{}); ok {
									log.Printf("[DEBUG] Converting filters from TypeList to TypeSet with %d items", len(filtersList))
									// Keep the data as a list in rawState and let Terraform's schema processing
									// convert it to TypeSet during normal resource loading. This avoids JSON serialization
									// issues with schema.Set objects that contain hash functions.
									filterGroupMap["filters"] = filtersList
									log.Printf("[DEBUG] Successfully prepared filters for TypeSet conversion")
								} else {
									log.Printf("[DEBUG] filters is not a list, skipping conversion")
								}
							} else {
								log.Printf("[DEBUG] No filters found in filter_group, skipping conversion")
							}
						}
					}
				}
			}
		}
	} else {
		log.Printf("[DEBUG] No spec found in state, skipping conversion")
	}

	return rawState, nil
}

// resourceFilterItemHash creates a hash for filter items in the TypeSet
func resourceFilterItemHash(v interface{}) int {
	var buf bytes.Buffer
	filter := v.(map[string]interface{})

	// Required fields - always include
	if key, ok := filter["key"].(string); ok {
		buf.WriteString(fmt.Sprintf("key:%s-", key))
	}

	if operator, ok := filter["operator"].(string); ok {
		buf.WriteString(fmt.Sprintf("operator:%s-", operator))
	}

	// Optional field with default
	if negation, ok := filter["negation"].(bool); ok {
		buf.WriteString(fmt.Sprintf("negation:%t-", negation))
	}

	// Handle values list - sort for deterministic hash
	if valuesRaw, ok := filter["values"]; ok && valuesRaw != nil {
		if valuesList, ok := valuesRaw.([]interface{}); ok {
			valuesStr := make([]string, len(valuesList))
			for i, v := range valuesList {
				if str, ok := v.(string); ok {
					valuesStr[i] = str
				}
			}
			sort.Strings(valuesStr)
			buf.WriteString(fmt.Sprintf("values:%s-", strings.Join(valuesStr, ",")))
		}
	}

	return int(hash(buf.String()))
}

// hash function (if not already available in the package)
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
