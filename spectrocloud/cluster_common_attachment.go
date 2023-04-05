package spectrocloud

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

var resourceAddonDeploymentCreatePendingStates = []string{
	"Node:NotReady",
	"Pack:Error",
	"PackPending",
	"Pack:NotReady",
	"Profile:NotAttached",
}

func waitForAddonDeployment(ctx context.Context, d *schema.ResourceData, cluster_uid string, profile_uid string, diags diag.Diagnostics, c *client.V1Client, state string) (diag.Diagnostics, bool) {
	cluster, err := c.GetCluster(cluster_uid)
	if err != nil {
		return diags, true
	}

	if _, found := cluster.Metadata.Labels["skip_packs"]; found {
		return diags, true
	}

	stateConf := &retry.StateChangeConf{
		Pending:    resourceAddonDeploymentCreatePendingStates,
		Target:     []string{"True"},
		Refresh:    resourceAddonDeploymentStateRefreshFunc(c, cluster_uid, profile_uid),
		Timeout:    d.Timeout(state) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err), true
	}
	return nil, false
}

func waitForAddonDeploymentCreation(ctx context.Context, d *schema.ResourceData, cluster_uid string, profile_uid string, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	return waitForAddonDeployment(ctx, d, cluster_uid, profile_uid, diags, c, schema.TimeoutCreate)
}

func waitForAddonDeploymentUpdate(ctx context.Context, d *schema.ResourceData, cluster_uid string, profile_uid string, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	return waitForAddonDeployment(ctx, d, cluster_uid, profile_uid, diags, c, schema.TimeoutUpdate)
}

func resourceAddonDeploymentStateRefreshFunc(c *client.V1Client, cluster_uid string, profile_uid string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := c.GetCluster(cluster_uid)
		if err != nil {
			return nil, "", err
		} else if cluster == nil {
			return nil, "Deleted", nil
		}

		// wait for nodes to be ready
		for _, node_condition := range cluster.Status.Conditions {
			if *node_condition.Status != "True" {
				log.Printf("Node state (%s): %s", cluster_uid, *node_condition.Status)
				return cluster, "Node:NotReady", nil
			}
		}

		// wait for profile to attach
		found := false
		for _, pack_status := range cluster.Status.Packs {
			if pack_status.ProfileUID == profile_uid {
				found = true
			}
		}

		if !found {
			return cluster, "Profile:NotAttached", nil
		}

		for _, pack_status := range cluster.Status.Packs {
			if pack_status.ProfileUID == profile_uid { // check only packs within this profile
				log.Printf("Pack state (%s): %s, %s", cluster_uid, pack_status.Name, *pack_status.Condition.Status)
				if *pack_status.Condition.Type == "Error" {
					return cluster, "Pack:Error", errors.New(pack_status.Condition.Message)
				}
				if *pack_status.Condition.Status != "True" || *pack_status.Condition.Type != "Ready" {
					return cluster, "Pack:NotReady", nil
				}
			}
		}

		return cluster, "True", nil
	}
}

func resourceAddonDeploymentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	clusterC, err := c.GetClusterClient()
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	cluster_uid := d.Get("cluster_uid").(string)
	cluster, err := c.GetCluster(cluster_uid)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		return diags
	}

	profile_uids := make([]string, 0)
	profileId, err := getClusterProfileUID(d.Id())
	if err != nil {
		return diags
	}
	profile_uids = append(profile_uids, profileId)

	if len(profile_uids) > 0 {
		err = c.DeleteAddonDeployment(clusterC, cluster_uid, &models.V1SpectroClusterProfilesDeleteEntity{
			ProfileUids: profile_uids,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
