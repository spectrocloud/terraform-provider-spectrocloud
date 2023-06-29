package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// read common fields like kubeconfig, tags, backup policy, scan policy, cluster_rbac_binding, namespaces
func readCommonFields(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) (diag.Diagnostics, bool) {
	kubecfg, err := c.GetClusterKubeConfig(d.Id())
	if err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("kubeconfig", kubecfg); err != nil {
		return diag.FromErr(err), true
	}

	if err := d.Set("tags", flattenTags(cluster.Metadata.Labels)); err != nil {
		return diag.FromErr(err), true
	}

	if policy, err := c.GetClusterBackupConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if policy != nil && policy.Spec.Config != nil {
		if err := d.Set("backup_policy", flattenBackupPolicy(policy.Spec.Config)); err != nil {
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
	if d.HasChanges("name", "tags") {
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

	if d.HasChange("location_config") {
		if err := updateLocationConfig(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	return diag.Diagnostics{}, false
}
