package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterConfigPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterConfigPolicyRead,
		Description: "Data source for retrieving information about a cluster config policy (maintenance policy).",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cluster config policy.",
			},
			"schedules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of maintenance schedules for the policy.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the upgrade schedule.",
						},
						"start_cron": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Cron expression for the start time of the schedule.",
						},
						"duration_hrs": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Specifies the time window in hours during which the system is allowed to start upgrades on eligible clusters.",
						},
					},
				},
			},
		},
	}
}

func dataSourceClusterConfigPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	policySummary, err := c.GetClusterConfigPolicyByName(name)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	policy, err := c.GetClusterConfigPolicy(policySummary.Metadata.UID)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	d.SetId(policy.Metadata.UID)

	if err := d.Set("name", policy.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if policy.Spec != nil {
		if err := d.Set("schedules", flattenClusterConfigPolicySchedules(policy.Spec.Schedules)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
