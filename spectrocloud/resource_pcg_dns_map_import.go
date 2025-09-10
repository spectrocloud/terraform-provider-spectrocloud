package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourcePrivateCloudGatewayDNSMapImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonPrivateCloudGatewayDNSMap(d, m)
	if err != nil {
		return nil, err
	}

	// Read all DNS map data to populate the state
	diags := resourcePCGDNSMapRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read PCG DNS map for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonPrivateCloudGatewayDNSMap(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// DNS maps are tenant-level resources, so use tenant context
	c := getV1ClientWithResourceContext(m, "tenant")

	// Parse the import ID to extract PCG ID and DNS map ID
	// Expected format: pcg_id:dns_map_id
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("DNS map import ID is required")
	}

	parts := strings.Split(importID, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format. Expected format: pcg_id:dns_map_id, got: %s", importID)
	}

	pcgID := parts[0]
	dnsMapID := parts[1]

	// Validate that the PCG exists
	pcg, err := c.GetPCGByID(pcgID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve private cloud gateway: %s", err)
	}
	if pcg == nil {
		return nil, fmt.Errorf("private cloud gateway with ID %s not found", pcgID)
	}

	// Validate that the DNS map exists
	dnsMap, err := c.GetVsphereDNSMap(dnsMapID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve DNS map: %s", err)
	}
	if dnsMap == nil {
		return nil, fmt.Errorf("DNS map with ID %s not found", dnsMapID)
	}

	// Set the required fields for the resource from the retrieved DNS map
	if err := d.Set("private_cloud_gateway_id", pcgID); err != nil {
		return nil, err
	}

	if dnsMap.Spec != nil {
		if dnsMap.Spec.DNSName != nil && *dnsMap.Spec.DNSName != "" {
			if err := d.Set("search_domain_name", *dnsMap.Spec.DNSName); err != nil {
				return nil, err
			}
		}

		if dnsMap.Spec.Datacenter != nil && *dnsMap.Spec.Datacenter != "" {
			if err := d.Set("data_center", *dnsMap.Spec.Datacenter); err != nil {
				return nil, err
			}
		}

		if dnsMap.Spec.Network != nil && *dnsMap.Spec.Network != "" {
			if err := d.Set("network", *dnsMap.Spec.Network); err != nil {
				return nil, err
			}
		}
	}

	// Set the ID to just the DNS map ID (the read function will use private_cloud_gateway_id)
	d.SetId(dnsMapID)

	return c, nil
}
