package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"log"
	"time"
)

var resourceClusterDeletePendingStates = []string{
	"Pending",
	"Provisioning",
	"Running",
	"Deleting",
	"Importing",
}
var resourceClusterCreatePendingStates = []string{
	"Unknown",
	"Pending",
	"Provisioning",
	"Importing",
}

func waitForClusterCreation(ctx context.Context, d *schema.ResourceData, uid string, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	d.SetId(uid)

	if _, found := toTags(d)["skip_completion"]; found {
		return diags, true
	}

	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterCreatePendingStates,
		Target:     []string{"Running"},
		Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err), true
	}
	return nil, false
}

//var resourceClusterUpdatePendingStates = []string{
//	"backing-up",
//	"modifying",
//	"resetting-master-credentials",
//	"upgrading",
//}
func waitForClusterDeletion(ctx context.Context, c *client.V1Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterDeletePendingStates,
		Target:     nil, // wait for deleted
		Refresh:    resourceClusterStateRefreshFunc(c, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func resourceClusterStateRefreshFunc(c *client.V1Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := c.GetCluster(id)
		if err != nil {
			return nil, "", err
		} else if cluster == nil {
			return nil, "Deleted", nil
		}

		state := cluster.Status.State
		log.Printf("Cluster state (%s): %s", id, state)

		return cluster, state, nil
	}
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	err := c.DeleteCluster(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := waitForClusterDeletion(ctx, c, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
