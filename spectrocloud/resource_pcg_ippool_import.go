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

	// Parse the import ID: pcg_id_or_name:ip_pool_id_or_name
	// Supports: pcg_id:ip_pool_id, pcg_name:ip_pool_name, pcg_id:ip_pool_name, pcg_name:ip_pool_id
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("IP pool import ID is required")
	}

	parts := strings.Split(importID, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format. Expected pcg_id_or_name:ip_pool_id_or_name, got: %s", importID)
	}

	pcgPart := parts[0]
	ipPoolPart := parts[1]

	// Resolve PCG: try by ID first, then by name
	pcgUID := ""
	pcg, err := c.GetPCGByID(pcgPart)
	if err == nil && pcg != nil && pcg.Metadata != nil && pcg.Metadata.UID != "" {
		pcgUID = pcg.Metadata.UID
	} else {
		pcg, err = c.GetPCGByName(pcgPart)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve private cloud gateway '%s': %w", pcgPart, err)
		}
		if pcg == nil || pcg.Metadata == nil || pcg.Metadata.UID == "" {
			return nil, fmt.Errorf("private cloud gateway not found: %s", pcgPart)
		}
		pcgUID = pcg.Metadata.UID
	}

	// Resolve IP pool: try by ID first, then by name
	ipPool, err := c.GetIPPool(pcgUID, ipPoolPart)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve IP pool: %w", err)
	}
	if ipPool == nil {
		ipPool, err = c.GetIPPoolByName(pcgUID, ipPoolPart)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve IP pool by name or id '%s': %w", ipPoolPart, err)
		}
		if ipPool == nil {
			return nil, fmt.Errorf("IP pool not found for name or id: %s", ipPoolPart)
		}
	}

	// Set the required fields for the resource
	if err := d.Set("private_cloud_gateway_id", pcgUID); err != nil {
		return nil, err
	}

	if ipPool.Metadata != nil {
		if err := d.Set("name", ipPool.Metadata.Name); err != nil {
			return nil, err
		}
	}

	// Set the network type based on the pool configuration
	networkType := "range" // default
	if ipPool.Spec != nil && ipPool.Spec.Pool != nil && len(ipPool.Spec.Pool.Subnet) > 0 {
		networkType = "subnet"
	}
	if err := d.Set("network_type", networkType); err != nil {
		return nil, err
	}

	// Set the ID to the IP pool UID
	if ipPool.Metadata != nil && ipPool.Metadata.UID != "" {
		d.SetId(ipPool.Metadata.UID)
	} else {
		return nil, fmt.Errorf("resolved IP pool has no UID")
	}

	return c, nil
}
