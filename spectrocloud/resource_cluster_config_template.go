package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceClusterConfigTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterConfigTemplateCreate,
		ReadContext:   resourceClusterConfigTemplateRead,
		UpdateContext: resourceClusterConfigTemplateUpdate,
		DeleteContext: resourceClusterConfigTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterConfigTemplateImport,
		},
		Description: "A resource for creating and managing cluster config templates.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cluster config template.",
			},
			"cloud_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cloud type for the cluster template. Examples: 'aws', 'azure', 'gcp', 'vsphere', etc.",
			},
			"profiles": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of cluster profile references.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "UID of the cluster profile.",
						},
					},
				},
			},
			"policies": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of policy references.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "UID of the policy.",
						},
						"kind": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Kind of the policy.",
						},
					},
				},
			},
		},
	}
}

func resourceClusterConfigTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	template := &models.V1ClusterTemplateEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1ClusterTemplateEntitySpec{
			CloudType: d.Get("cloud_type").(string),
			Profiles:  expandClusterTemplateProfiles(d.Get("profiles").([]interface{})),
			Policies:  expandClusterTemplatePolicies(d.Get("policies").([]interface{})),
		},
	}

	uid, err := c.CreateClusterConfigTemplate(template)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*uid.UID)
	return resourceClusterConfigTemplateRead(ctx, d, m)
}

func resourceClusterConfigTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	uid := d.Id()

	template, err := c.GetClusterConfigTemplate(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	if err := d.Set("name", template.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if template.Spec != nil {
		if err := d.Set("cloud_type", template.Spec.CloudType); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("profiles", flattenClusterTemplateProfiles(template.Spec.Profiles)); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("policies", flattenClusterTemplatePolicies(template.Spec.Policies)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceClusterConfigTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	metadata := &models.V1ObjectMetaInputEntitySchema{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name: d.Get("name").(string),
		},
	}

	err := c.UpdateClusterConfigTemplate(d.Id(), metadata)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceClusterConfigTemplateRead(ctx, d, m)
}

func resourceClusterConfigTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	err := c.DeleteClusterConfigTemplate(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceClusterConfigTemplateImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// The ID passed in is the UID
	d.SetId(d.Id())

	diags := resourceClusterConfigTemplateRead(ctx, d, m)
	if diags.HasError() {
		return nil, diags[0].Validate()
	}

	return []*schema.ResourceData{d}, nil
}

// Helper functions for expanding and flattening

func expandClusterTemplateProfiles(profiles []interface{}) []*models.V1ClusterTemplateProfile {
	if len(profiles) == 0 {
		return nil
	}

	result := make([]*models.V1ClusterTemplateProfile, len(profiles))
	for i, profile := range profiles {
		p := profile.(map[string]interface{})
		result[i] = &models.V1ClusterTemplateProfile{
			UID: p["uid"].(string),
		}
	}

	return result
}

func expandClusterTemplatePolicies(policies []interface{}) []*models.V1PolicyRef {
	if len(policies) == 0 {
		return nil
	}

	result := make([]*models.V1PolicyRef, len(policies))
	for i, policy := range policies {
		p := policy.(map[string]interface{})
		result[i] = &models.V1PolicyRef{
			UID:  p["uid"].(string),
			Kind: p["kind"].(string),
		}
	}

	return result
}

func flattenClusterTemplateProfiles(profiles []*models.V1ClusterTemplateProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(profiles))
	for i, profile := range profiles {
		result[i] = map[string]interface{}{
			"uid": profile.UID,
		}
	}

	return result
}

func flattenClusterTemplatePolicies(policies []*models.V1PolicyRef) []interface{} {
	if policies == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(policies))
	for i, policy := range policies {
		result[i] = map[string]interface{}{
			"uid":  policy.UID,
			"kind": policy.Kind,
		}
	}

	return result
}
