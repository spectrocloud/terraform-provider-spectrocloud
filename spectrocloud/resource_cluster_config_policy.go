package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceClusterConfigPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterConfigPolicyCreate,
		ReadContext:   resourceClusterConfigPolicyRead,
		UpdateContext: resourceClusterConfigPolicyUpdate,
		DeleteContext: resourceClusterConfigPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterConfigPolicyImport,
		},
		Description: "A resource for creating and managing cluster config policies (maintenance policies).",

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
				Description: "The name of the cluster config policy.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description: "The context of the cluster config policy. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Assign tags to the cluster config policy. Tags can be in the format `key:value` or just `key`.",
			},
			"schedules": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of maintenance schedules for the policy.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the upgrade schedule.",
						},
						"start_cron": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Cron expression for the start time of the schedule.",
						},
						"duration_hrs": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Specifies the time window in hours during which the system is allowed to start upgrades on eligible clusters. Valid range: 1-24.",
						},
					},
				},
			},
		},
	}
}

func resourceClusterConfigPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))

	policy := &models.V1SpcPolicyEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			Labels: toTags(d),
		},
		Spec: &models.V1SpcPolicySpec{
			Schedules: expandClusterConfigPolicySchedules(d.Get("schedules").([]interface{})),
		},
	}

	uid, err := c.CreateClusterConfigPolicy(policy)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*uid.UID)
	return resourceClusterConfigPolicyRead(ctx, d, m)
}

func resourceClusterConfigPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))
	var diags diag.Diagnostics
	uid := d.Id()

	policy, err := c.GetClusterConfigPolicy(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	if err := d.Set("name", policy.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tags", flattenTags(policy.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	if policy.Spec != nil {
		if err := d.Set("schedules", flattenClusterConfigPolicySchedules(policy.Spec.Schedules)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceClusterConfigPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))

	policy := &models.V1SpcPolicyEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			Labels: toTags(d),
		},
		Spec: &models.V1SpcPolicySpec{
			Schedules: expandClusterConfigPolicySchedules(d.Get("schedules").([]interface{})),
		},
	}

	err := c.UpdateClusterConfigPolicy(d.Id(), policy)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceClusterConfigPolicyRead(ctx, d, m)
}

func resourceClusterConfigPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))

	err := c.DeleteClusterConfigPolicy(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceClusterConfigPolicyImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// The ID passed in is the UID
	d.SetId(d.Id())

	diags := resourceClusterConfigPolicyRead(ctx, d, m)
	if diags.HasError() {
		return nil, diags[0].Validate()
	}

	return []*schema.ResourceData{d}, nil
}

// Helper functions for expanding and flattening

func expandClusterConfigPolicySchedules(schedules []interface{}) []*models.V1Schedule {
	if len(schedules) == 0 {
		return nil
	}

	result := make([]*models.V1Schedule, len(schedules))
	for i, schedule := range schedules {
		s := schedule.(map[string]interface{})
		name := s["name"].(string)
		startCron := s["start_cron"].(string)
		durationHrs := int64(s["duration_hrs"].(int))

		result[i] = &models.V1Schedule{
			Name:        &name,
			StartCron:   &startCron,
			DurationHrs: &durationHrs,
		}
	}

	return result
}

func flattenClusterConfigPolicySchedules(schedules []*models.V1Schedule) []interface{} {
	if schedules == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(schedules))
	for i, schedule := range schedules {
		m := map[string]interface{}{}
		if schedule.Name != nil {
			m["name"] = *schedule.Name
		}
		if schedule.StartCron != nil {
			m["start_cron"] = *schedule.StartCron
		}
		if schedule.DurationHrs != nil {
			m["duration_hrs"] = int(*schedule.DurationHrs)
		}
		result[i] = m
	}

	return result
}
