package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

func resourceClusterVirtualImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "project")
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("virtual cluster import ID or name is required")
	}

	var cluster *models.V1SpectroCluster
	var err error

	// Try by UID first
	cluster, err = c.GetCluster(importID)
	if err != nil && !herr.IsNotFound(err) {
		return nil, fmt.Errorf("could not retrieve virtual cluster for import: %s", err)
	}
	if err == nil && cluster != nil {
		d.SetId(cluster.Metadata.UID)
	} else {
		// Try by name (virtual clusters use GetClusterByName(..., true))
		cluster, err = c.GetClusterByName(importID, true)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve virtual cluster by ID or name '%s': %s", importID, err)
		}
		if cluster == nil {
			return nil, fmt.Errorf("virtual cluster with ID or name '%s' not found", importID)
		}
		d.SetId(cluster.Metadata.UID)
	}

	// Set the cluster name from the retrieved cluster
	if err := d.Set("name", cluster.Metadata.Name); err != nil {
		return nil, err
	}
	// Set the context to project as default for import
	if err := d.Set("context", "project"); err != nil {
		return nil, err
	}
	if err := d.Set("cluster_group_uid", cluster.Spec.ClusterConfig.HostClusterConfig.ClusterGroup.UID); err != nil {
		return nil, err
	}
	// setting up default settings for import
	if err := d.Set("host_cluster_uid", cluster.Spec.ClusterConfig.HostClusterConfig.HostCluster.UID); err != nil {
		return nil, err
	}
	if err := d.Set("apply_setting", "DownloadAndInstall"); err != nil {
		return nil, err
	}
	if err := d.Set("force_delete", false); err != nil {
		return nil, err
	}
	if err := d.Set("force_delete_delay", 20); err != nil {
		return nil, err
	}
	if err := d.Set("skip_completion", false); err != nil {
		return nil, err
	}
	if err := d.Set("os_patch_on_boot", false); err != nil {
		return nil, err
	}
	if err := d.Set("pause_cluster", false); err != nil {
		return nil, err
	}
	// Read all cluster data to populate the state
	diags := resourceClusterVirtualRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read virtual cluster for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}
