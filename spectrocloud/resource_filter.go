package spectrocloud

import (
	"context"
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
		Description:   "A resource for creating and managing filters.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
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
										Type:        schema.TypeList,
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
