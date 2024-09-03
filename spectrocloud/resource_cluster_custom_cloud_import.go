package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func resourceClusterCustomImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clusterID, scope, customCloudName, err := ParseResourceCustomCloudImportID(d)
	if err != nil {
		return nil, err
	}
	d.SetId(clusterID + ":" + scope)
	_ = d.Set("cloud", customCloudName)
	c, err := GetCommonCluster(d, m)
	if err != nil {
		return nil, err
	}
	diags := resourceClusterCustomCloudRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// cluster profile and common default cluster attribute is get set here
	err = flattenCommonAttributeForCustomClusterImport(c, d)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func ParseResourceCustomCloudImportID(d *schema.ResourceData) (string, string, string, error) {
	// Example: If the ID is a combination of ClusterId, then name of context/scope: `project` or `tenant` and then cloud type
	// and if scope is then followed by projectID  "cluster456:project:nutanix" or "cluster456:tenant:oracle"
	parts := strings.Split(d.Id(), ":")

	scope := "invalid"
	clusterID := ""
	customCloudName := ""
	if len(parts) == 3 && (parts[1] == "tenant" || parts[1] == "project") {
		clusterID, scope, customCloudName = parts[0], parts[1], parts[2]
	}
	if scope == "invalid" {
		return "", "", "", fmt.Errorf("invalid cluster ID format specified for import custom cloud %s, Ex: it should cluster_id:context:custom_cloud_name, `cluster456:project:nutanix`", d.Id())
	}
	return clusterID, scope, customCloudName, nil
}
