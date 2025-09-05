package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourcePrivateCloudGatewayIpPoolImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonPrivateCloudGatewayIpPool(d, m)
	if err != nil {
		return nil, err
	}

	// Read all IP pool data to populate the state
	diags := resourceIpPoolRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read IP pool for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonPrivateCloudGatewayIpPool(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// IP pools are tenant-level resources, so use tenant context
	c := getV1ClientWithResourceContext(m, "tenant")

	// Parse the import ID to extract PCG ID and IP pool ID
	// Expected format: pcg_id:ippool_id
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("IP pool import ID is required")
	}

	parts := strings.Split(importID, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format. Expected format: pcg_id:ippool_id, got: %s", importID)
	}

	pcgID := parts[0]
	ipPoolID := parts[1]

	// Validate that the PCG exists
	pcg, err := c.GetPCGByID(pcgID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve private cloud gateway: %s", err)
	}
	if pcg == nil {
		return nil, fmt.Errorf("private cloud gateway with ID %s not found", pcgID)
	}

	// Validate that the IP pool exists within the PCG
	ipPool, err := c.GetIPPool(pcgID, ipPoolID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve IP pool: %s", err)
	}
	if ipPool == nil {
		return nil, fmt.Errorf("IP pool with ID %s not found in PCG %s", ipPoolID, pcgID)
	}

	// Set the required fields for the resource
	if err := d.Set("private_cloud_gateway_id", pcgID); err != nil {
		return nil, err
	}

	if err := d.Set("name", ipPool.Metadata.Name); err != nil {
		return nil, err
	}

	// Set the network type based on the pool configuration
	networkType := "range" // default
	if ipPool.Spec != nil && ipPool.Spec.Pool != nil && len(ipPool.Spec.Pool.Subnet) > 0 {
		networkType = "subnet"
	}
	if err := d.Set("network_type", networkType); err != nil {
		return nil, err
	}

	// Set the ID to just the IP pool ID (the read function will use private_cloud_gateway_id)
	d.SetId(ipPoolID)

	return c, nil
}
