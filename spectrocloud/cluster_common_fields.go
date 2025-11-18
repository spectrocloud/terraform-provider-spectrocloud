package spectrocloud

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// read common fields like kubeconfig, tags, backup policy, scan policy, cluster_rbac_binding, namespaces
func readCommonFields(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) (diag.Diagnostics, bool) {
	//ClusterContext := "project"
	//if cluster.Metadata.Annotations["scope"] != "" {
	//	ClusterContext = cluster.Metadata.Annotations["scope"]
	//}
	kubecfg, err := c.GetClusterClientKubeConfig(d.Id())
	if err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("kubeconfig", kubecfg); err != nil {
		return diag.FromErr(err), true
	}
	// When the current repave state is pending, we set the review_repave_state to Pending, For indicate the system change.
	if _, ok := d.GetOk("review_repave_state"); ok {
		// We are adding this check to handle virtual cluster scenario. virtual cluster doesn't have support for `review_repave_state`
		if err := d.Set("review_repave_state", cluster.Status.Repave.State); err != nil {
			return diag.FromErr(err), true
		}
	}
	adminKubeConfig, err := c.GetClusterAdminKubeConfig(d.Id())
	if err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("admin_kube_config", adminKubeConfig); err != nil {
		return diag.FromErr(err), true
	}

	if err := d.Set("tags", flattenTags(cluster.Metadata.Labels)); err != nil {
		return diag.FromErr(err), true
	}

	if policy, err := c.GetClusterBackupConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if policy != nil && policy.Spec.Config != nil {
		if err := d.Set("backup_policy", flattenBackupPolicy(policy.Spec.Config, d)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if policy, err := c.GetClusterScanConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if policy != nil && policy.Spec.DriverSpec != nil {
		if err := d.Set("scan_policy", flattenScanPolicy(policy.Spec.DriverSpec)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if rbac, err := c.GetClusterRbacConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if rbac != nil && rbac.Items != nil {
		if err := d.Set("cluster_rbac_binding", flattenClusterRBAC(rbac.Items)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if namespace, err := c.GetClusterNamespaceConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if namespace != nil && namespace.Items != nil {
		if err := d.Set("namespaces", flattenClusterNamespaces(namespace.Items)); err != nil {
			return diag.FromErr(err), true
		}
	}

	clusterAdditionalMeta := cluster.Spec.ClusterConfig.ClusterMetaAttribute
	if clusterAdditionalMeta != "" {
		// We are adding this check to handle virtual cluster scenario. virtual cluster doesn't have support for `cluster_meta_attribute`
		if _, ok := d.GetOk("cluster_meta_attribute"); ok {
			err := d.Set("cluster_meta_attribute", clusterAdditionalMeta)
			if err != nil {
				return diag.FromErr(err), true
			}
		}
	}

	hostConfig := cluster.Spec.ClusterConfig.HostClusterConfig
	if hostConfig != nil && *hostConfig.IsHostCluster {
		flattenHostConfig := flattenHostConfig(hostConfig)
		if len(flattenHostConfig) > 0 {
			if err := d.Set("host_config", flattenHostConfig); err != nil {
				return diag.FromErr(err), true
			}
		}
	}

	if _, ok := d.GetOk("pause_agent_upgrades"); ok {
		if err := d.Set("pause_agent_upgrades", getSpectroComponentsUpgrade(cluster)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if _, ok := d.GetOk("review_repave_state"); ok {
		if err := d.Set("review_repave_state", cluster.Status.Repave.State); err != nil {
			return diag.FromErr(err), true
		}
	}

	//clusterContext := d.Get("context").(string)

	if clusterStatus, err := c.GetClusterWithoutStatus(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if clusterStatus != nil && clusterStatus.Status != nil && clusterStatus.Status.Location != nil {
		if err := d.Set("location_config", flattenLocationConfig(clusterStatus.Status.Location)); err != nil {
			return diag.FromErr(err), true
		}
	}

	return diag.Diagnostics{}, false
}

func getSpectroComponentsUpgrade(cluster *models.V1SpectroCluster) string {
	if cluster.Metadata.Annotations != nil {
		clusterAnnotation := cluster.Metadata.Annotations
		if v, ok := clusterAnnotation["spectroComponentsUpgradeForbidden"]; ok {
			if v == "true" {
				return "lock"
			}
			return "unlock"
		}
	}
	return "unlock"
}

// update common fields like namespaces, cluster_rbac_binding, cluster_profile, backup_policy, scan_policy
func updateCommonFields(d *schema.ResourceData, c *client.V1Client) (diag.Diagnostics, bool) {
	if d.HasChanges("name", "tags", "description", "tags_map") {
		if err := updateClusterMetadata(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("namespaces") {
		if err := updateClusterNamespaces(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("cluster_rbac_binding") {
		if err := updateClusterRBAC(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("os_patch_on_boot") || d.HasChange("os_patch_schedule") || d.HasChange("os_patch_after") {
		if err := updateClusterOsPatchConfig(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	// Handle cluster_template changes separately using variables API (doesn't trigger full cluster update)
	if d.HasChange("cluster_template") {
		if err := updateClusterTemplateVariables(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	// Handle cluster_profile changes using the existing profile update flow
	if d.HasChanges("cluster_profile", "packs", "manifests") {
		if err := updateProfiles(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("backup_policy") {
		if err := updateBackupPolicy(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("scan_policy") {
		if err := updateScanPolicy(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("host_config") {
		if err := updateHostConfig(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("pause_agent_upgrades") {
		if err := updateAgentUpgradeSetting(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("cluster_meta_attribute") {
		if err := updateClusterAdditionalMetadata(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("location_config") {
		if err := updateLocationConfig(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	return diag.Diagnostics{}, false
}

func validateSystemRepaveApproval(d *schema.ResourceData, c *client.V1Client) error {
	approveClusterRepave := d.Get("review_repave_state").(string)
	//context := d.Get("context").(string)
	cluster, err := c.GetCluster(d.Id())
	if err != nil {
		return err
	}
	if cluster == nil {
		return nil
	}
	if cluster.Status.Repave.State != nil {
		if *cluster.Status.Repave.State == models.V1ClusterRepaveStatePending {
			if approveClusterRepave == "Approved" {
				err := c.ApproveClusterRepave(d.Id())
				if err != nil {
					return err
				}
				cluster, err := c.GetCluster(d.Id())
				if err != nil {
					return err
				}
				if *cluster.Status.Repave.State == models.V1ClusterRepaveStateApproved {
					return nil
				} else {
					err = errors.New("repave cluster is not approved - cluster repave state is still not approved. Please set `review_repave_state` to `Approved` to approve the repave operation on the cluster")
					return err
				}

			} else {
				reasons, err := c.GetRepaveReasons(d.Id())
				if err != nil {
					return err
				}
				err = errors.New("cluster repave state is pending. \nDue to the following reasons -  \n" + strings.Join(reasons, "\n") + "\nKindly verify the cluster and set `review_repave_state` to `Approved` to continue the repave operation and day 2 operation on the cluster.")
				return err
			}
		}
	}

	return nil
}

func validateReviewRepaveValue(val interface{}, key string) (warns []string, errs []error) {
	repaveValue := val.(string)
	validStatuses := map[string]bool{
		"":         true,
		"Approved": true,
		"Pending":  true,
	}
	if repaveValue == "Approved" {
		warning := []string{"Review Repave Value Warning:",
			"Setting `review_repave_state` to `Approved` will authorize the palette to repave the cluster if any system repave is in the pending state. Please exercise caution when using `review_repave_state` attribute."}
		warns = append(warns, strings.Join(warning, "\n"))
	}
	if _, ok := validStatuses[repaveValue]; !ok {
		errs = append(errs, fmt.Errorf("expected review_repave_state to be one of [``, `Pending`, `Approved`], got %s", repaveValue))
	}
	return warns, errs
}
