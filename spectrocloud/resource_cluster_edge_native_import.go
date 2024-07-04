package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterEdgeNativeImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	scope, _, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	c := GetResourceLevelV1Client(m, scope)

	err = GetCommonCluster(d, c)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterEdgeNativeRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// cluster profile and common default cluster attribute is get set here
	err = flattenCommonAttributeForClusterImport(c, d)
	if err != nil {
		return nil, err
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

func ParseResourceID(d *schema.ResourceData) (string, string, error) {
	// d.Id() will contain the ID of the resource to import. This ID is provided by the user
	// during the import command, and should be parsed to find the existing resource.
	// Example: `terraform import spectrocloud_cluster.my_cluster [id]`

	// Parse the ID to find the existing resource. This might involve making API requests
	// to your infrastructure with the client `c`.
	// Example: If the ID is a combination of ClusterId, then name of context/scope: `project` or `tenant`
	// and if scope is then followed by projectID  "cluster456:project" or "cluster456:tenant"
	parts := strings.Split(d.Id(), ":")
	// if 2 parts - last part should be `tenant`
	scope := "invalid"
	clusterID := ""
	if len(parts) == 2 && (parts[1] == "tenant" || parts[1] == "project") {
		clusterID, scope = parts[0], parts[1]
	}
	if scope == "invalid" {
		return "", "", fmt.Errorf("invalid cluster ID format specified for import %s", d.Id())
	}
	return scope, clusterID, nil
}
