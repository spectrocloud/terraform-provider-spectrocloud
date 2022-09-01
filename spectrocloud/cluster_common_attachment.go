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

var resourceAddonDeploymentCreatePendingStates = []string{
	"Node:False",
	"Pack:False",
}

func waitForAddonDeploymentCreation(ctx context.Context, d *schema.ResourceData, cluster *models.V1SpectroCluster, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	if _, found := toTags(d)["skip_packs"]; found {
		return diags, true
	}

	stateConf := &resource.StateChangeConf{
		Pending:    resourceAddonDeploymentCreatePendingStates,
		Target:     []string{"True"},
		Refresh:    resourceAddonDeploymentStateRefreshFunc(c, cluster.Metadata.UID),
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

func resourceAddonDeploymentStateRefreshFunc(c *client.V1Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := c.GetCluster(id)
		if err != nil {
			return nil, "", err
		} else if cluster == nil {
			return nil, "Deleted", nil
		}

		for _, pack_status := range cluster.Status.Packs {
			if *pack_status.Condition.Status != "True" {
				log.Printf("Pack state (%s): %s, %s", id, pack_status.Name, *pack_status.Condition.Status)
				return cluster, "Pack:" + *pack_status.Condition.Status, nil
			}
		}
		for _, node_condition := range cluster.Status.Conditions {
			if *node_condition.Status != "True" {
				log.Printf("Node state (%s): %s", id, *node_condition.Status)
				return cluster, "Node:" + *node_condition.Status, nil
			}
		}

		return cluster, "True", nil
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

	return diags
}
