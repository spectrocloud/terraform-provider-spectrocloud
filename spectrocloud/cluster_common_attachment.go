package spectrocloud

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"log"
	"time"
)

var resourceAddonDeploymentCreatePendingStates = []string{
	"Node:NotReady",
	"Pack:Error",
	"PackPending",
	"Pack:NotReady",
	"Profile:NotAttached",
}

func waitForAddonDeploymentCreation(ctx context.Context, d *schema.ResourceData, cluster_uid string, profile_uid string, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	cluster, err := c.GetCluster(cluster_uid)
	if err != nil {
		return diags, true
	}

	if _, found := cluster.Metadata.Labels["skip_packs"]; found {
		return diags, true
	}

	stateConf := &resource.StateChangeConf{
		Pending:    resourceAddonDeploymentCreatePendingStates,
		Target:     []string{"True"},
		Refresh:    resourceAddonDeploymentStateRefreshFunc(c, cluster_uid, profile_uid),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
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

func resourceAddonDeploymentStateRefreshFunc(c *client.V1Client, cluster_uid string, profile_uid string) resource.StateRefreshFunc {
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

	var diags diag.Diagnostics
	cluster_uid := d.Get("cluster_uid").(string)
	cluster, err := c.GetCluster(cluster_uid)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		return diags
	}

	profile_uids := make([]string, 0)
	profiles := d.Get("cluster_profile").([]interface{})
	if len(profiles) > 0 {
		for _, profile := range profiles {
			p := profile.(map[string]interface{})
			if isProfileAttached(cluster, p["id"].(string)) {
				profile_uids = append(profile_uids, p["id"].(string))
			}
		}
	}

	if len(profile_uids) > 0 {
		err = c.DeleteAddonDeployment(cluster_uid, &models.V1SpectroClusterProfilesDeleteEntity{
			ProfileUids: profile_uids,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
