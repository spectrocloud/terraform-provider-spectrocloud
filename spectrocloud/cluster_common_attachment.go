package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"log"
	"time"
)

var resourceAddonDeploymentDeletePendingStates = []string{
	"Pending",
	"Provisioning",
	"Running",
	"Deleting",
	"Importing",
}
var resourceAddonDeploymentCreatePendingStates = []string{
	"Unknown",
	"Pending",
	"Provisioning",
	"Importing",
}

func waitForAddonDeploymentCreation(ctx context.Context, d *schema.ResourceData, cluster *models.V1SpectroCluster, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	if _, found := toTags(d)["wait_apply"]; found {
		return diags, true
	}

	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterCreatePendingStates,
		Target:     []string{"Running"},
		Refresh:    resourceAddonDeploymentStateRefreshFunc(c, d.Id()),
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
func waitForAddonDeploymentDeletion(ctx context.Context, c *client.V1Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterDeletePendingStates,
		Target:     nil, // wait for deleted
		Refresh:    resourceAddonDeploymentStateRefreshFunc(c, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func resourceAddonDeploymentStateRefreshFunc(c *client.V1Client, id string) resource.StateRefreshFunc {
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

func resourceAddonDeploymentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics
	profile_uids := make([]string, 0)
	profiles := d.Get("cluster_profile").([]interface{})
	if len(profiles) > 0 {
		for _, profile := range profiles {
			p := profile.(map[string]interface{})
			profile_uids = append(profile_uids, p["id"].(string))
		}
	}

	err := c.DeleteAddonDeploymentValues(d.Get("cluster_uid").(string), &models.V1SpectroClusterProfilesDeleteEntity{
		ProfileUids: profile_uids,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err := waitForAddonDeploymentDeletion(ctx, c, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
