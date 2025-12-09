package spectrocloud

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

var resourceAddonDeploymentCreatePendingStates = []string{
	"Node:NotReady",
	"Pack:Error",
	"PackPending",
	"Pack:NotReady",
	"Profile:NotAttached",
}

func waitForAddonDeployment(ctx context.Context, d *schema.ResourceData, cl models.V1SpectroCluster, profile_uid string, diags diag.Diagnostics, c *client.V1Client, state string) (diag.Diagnostics, bool) {
	cluster, err := c.GetCluster(cl.Metadata.UID)
	if err != nil {
		return diags, true
	}

	if _, found := cluster.Metadata.Labels["skip_packs"]; found {
		return diags, true
	}

	stateConf := &retry.StateChangeConf{
		Pending:    resourceAddonDeploymentCreatePendingStates,
		Target:     []string{"True"},
		Refresh:    resourceAddonDeploymentStateRefreshFunc(c, *cluster, profile_uid),
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

func waitForAddonDeploymentCreation(ctx context.Context, d *schema.ResourceData, cluster models.V1SpectroCluster, profile_uid string, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	return waitForAddonDeployment(ctx, d, cluster, profile_uid, diags, c, schema.TimeoutCreate)
}

func waitForAddonDeploymentUpdate(ctx context.Context, d *schema.ResourceData, cluster models.V1SpectroCluster, profile_uid string, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	return waitForAddonDeployment(ctx, d, cluster, profile_uid, diags, c, schema.TimeoutUpdate)
}

func resourceAddonDeploymentStateRefreshFunc(c *client.V1Client, cluster models.V1SpectroCluster, profile_uid string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := c.GetCluster(cluster.Metadata.UID)
		if err != nil {
			return nil, "", err
		} else if cluster == nil {
			return nil, "Deleted", nil
		}

		// wait for nodes to be ready
		for _, node_condition := range cluster.Status.Conditions {
			if *node_condition.Status != "True" {
				log.Printf("Node state (%s): %s", cluster.Metadata.UID, *node_condition.Status)
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
				log.Printf("Pack state (%s): %s, %s", cluster.Metadata.UID, pack_status.Name, *pack_status.Condition.Status)
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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics
	clusterUid := d.Get("cluster_uid").(string)
	cluster, err := c.GetCluster(clusterUid)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		return diags
	}

	// FIX: Read ALL profiles from state, not just the one in resource ID
	// The resource can manage multiple profiles via cluster_profile blocks
	profile_uids := make([]string, 0)
	profile_names := make(map[string]bool) // Track profile names from state
	stateProfilesRaw := d.Get("cluster_profile")
	if stateProfilesRaw != nil {
		if stateProfilesList, ok := stateProfilesRaw.([]interface{}); ok {
			log.Printf("Delete: Found %d profiles in state", len(stateProfilesList))
			for _, profileRaw := range stateProfilesList {
				profile := profileRaw.(map[string]interface{})
				if id, ok := profile["id"].(string); ok && id != "" {
					profile_uids = append(profile_uids, id)
					log.Printf("Delete: Adding profile UID from state: %s", id)
					// Also get profile name for matching
					if profileDef, err := c.GetClusterProfile(id); err == nil && profileDef != nil && profileDef.Metadata != nil {
						profile_names[profileDef.Metadata.Name] = true
						log.Printf("Delete: Profile UID %s has name: %s", id, profileDef.Metadata.Name)
					}
				}
			}
		}
	} else {
		log.Printf("Delete: No profiles found in state (stateProfilesRaw is nil)")
	}

	// CRITICAL FIX: Also find add-on profiles on cluster that match by name
	// This handles cases where profile UID changed (version update) or state is incomplete
	// We'll delete all add-on profiles that match the names of profiles in state
	if len(profile_names) > 0 {
		log.Printf("Delete: Checking cluster for add-on profiles matching %d profile names from state", len(profile_names))
		for _, templateProfile := range cluster.Spec.ClusterProfileTemplates {
			if templateProfile != nil && templateProfile.Name != "" {
				// Check if this profile name matches any profile in state
				if profile_names[templateProfile.Name] {
					// Verify it's an add-on profile
					profileDef, err := c.GetClusterProfile(templateProfile.UID)
					if err == nil && profileDef != nil && profileDef.Spec != nil && profileDef.Spec.Published != nil {
						if string(profileDef.Spec.Published.Type) == string(models.V1ProfileTypeAddDashOn) {
							// Add to deletion list if not already there
							alreadyInList := false
							for _, uid := range profile_uids {
								if uid == templateProfile.UID {
									alreadyInList = true
									break
								}
							}
							if !alreadyInList {
								profile_uids = append(profile_uids, templateProfile.UID)
								log.Printf("Delete: Found add-on profile %s (UID: %s) on cluster matching state profile name, will be deleted", templateProfile.Name, templateProfile.UID)
							}
						}
					}
				}
			}
		}
	}

	// Fallback: If state doesn't have profiles, try to get from resource ID
	if len(profile_uids) == 0 {
		log.Printf("Delete: No profiles from state, trying resource ID fallback")
		profileId, err := getClusterProfileUID(d.Id())
		if err == nil && profileId != "" {
			profile_uids = append(profile_uids, profileId)
			log.Printf("Delete: Using profile UID from resource ID: %s", profileId)
		}
	}

	if len(profile_uids) > 0 {
		log.Printf("Delete: Deleting %d add-on profiles: %v", len(profile_uids), profile_uids)
		err = c.DeleteAddonDeployment(clusterUid, &models.V1SpectroClusterProfilesDeleteEntity{
			ProfileUids: profile_uids,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		log.Printf("Delete: No profiles to delete (profile_uids is empty)")
	}

	return diags
}
