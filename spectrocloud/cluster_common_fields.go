package spectrocloud

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"strings"
)

// read common fields like kubeconfig, tags, backup policy, scan policy, cluster_rbac_binding, namespaces
func readCommonFields(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) (diag.Diagnostics, bool) {
	ClusterContext := "project"
	if cluster.Metadata.Annotations["scope"] != "" {
		ClusterContext = cluster.Metadata.Annotations["scope"]
	}
	kubecfg, err := c.GetClusterKubeConfig(d.Id(), ClusterContext)
	if err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("kubeconfig", kubecfg); err != nil {
		return diag.FromErr(err), true
	}

	adminKubeConfig, err := c.GetClusterAdminKubeConfig(d.Id(), ClusterContext)
	if err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("admin_kube_config", adminKubeConfig); err != nil {
		return diag.FromErr(err), true
	}

	if err := d.Set("tags", flattenTags(cluster.Metadata.Labels)); err != nil {
		return diag.FromErr(err), true
	}

	if policy, err := c.GetClusterBackupConfig(d.Id(), ClusterContext); err != nil {
		return diag.FromErr(err), true
	} else if policy != nil && policy.Spec.Config != nil {
		if err := d.Set("backup_policy", flattenBackupPolicy(policy.Spec.Config)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if policy, err := c.GetClusterScanConfig(d.Id(), ClusterContext); err != nil {
		return diag.FromErr(err), true
	} else if policy != nil && policy.Spec.DriverSpec != nil {
		if err := d.Set("scan_policy", flattenScanPolicy(policy.Spec.DriverSpec)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if rbac, err := c.GetClusterRbacConfig(d.Id(), ClusterContext); err != nil {
		return diag.FromErr(err), true
	} else if rbac != nil && rbac.Items != nil {
		if err := d.Set("cluster_rbac_binding", flattenClusterRBAC(rbac.Items)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if namespace, err := c.GetClusterNamespaceConfig(d.Id(), ClusterContext); err != nil {
		return diag.FromErr(err), true
	} else if namespace != nil && namespace.Items != nil {
		if err := d.Set("namespaces", flattenClusterNamespaces(namespace.Items)); err != nil {
			return diag.FromErr(err), true
		}
	}

	clusterAdditionalMeta := cluster.Spec.ClusterConfig.ClusterMetaAttribute
	if clusterAdditionalMeta != "" {
		err := d.Set("cluster_meta_attribute", clusterAdditionalMeta)
		if err != nil {
			return diag.FromErr(err), true
		}
	}
	repaveState := cluster.Status.Repave.State
	if repaveState != "" {
		err := d.Set("repave_state", repaveState)
		if err != nil {
			return diag.FromErr(err), true
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

	clusterContext := d.Get("context").(string)

	if clusterStatus, err := c.GetClusterWithoutStatus(clusterContext, d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if clusterStatus != nil && clusterStatus.Status != nil && clusterStatus.Status.Location != nil {
		if err := d.Set("location_config", flattenLocationConfig(clusterStatus.Status.Location)); err != nil {
			return diag.FromErr(err), true
		}
	}

	return diag.Diagnostics{}, false
}

// update common fields like namespaces, cluster_rbac_binding, cluster_profile, backup_policy, scan_policy
func updateCommonFields(d *schema.ResourceData, c *client.V1Client) (diag.Diagnostics, bool) {
	if d.HasChanges("name", "tags", "description") {
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
	approveClusterRepave := d.Get("approve_system_repave").(bool)
	context := d.Get("context").(string)
	cluster, err := c.GetCluster(context, d.Id())
	if err != nil {
		return err
	}
	if cluster.Status.Repave.State == "Pending" {
		if approveClusterRepave {
			err := c.ApproveClusterRepave(context, d.Id())
			if err != nil {
				return err
			}
			cluster, err := c.GetCluster(context, d.Id())
			if err != nil {
				return err
			}
			if cluster.Status.Repave.State == "Approved" {
				return nil
			} else {
				err = errors.New("repave cluster is not approved - cluster repave state is still not approved. Please set `approve_system_repave` to `true` to approve the repave operation on the cluster")
				return err
			}

		} else {
			reasons, err := c.GetRepaveReasons(context, d.Id())
			if err != nil {
				return err
			}
			err = errors.New("cluster repave state is pending. \nDue to the following reasons -  \n" + strings.Join(reasons, "\n") + "\nKindly verify the cluster and set `approve_system_repave` to `true` to continue the repave operation and day 2 operation on the cluster.")
			return err
		}
	}
	return nil
}
