package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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

	// Parse the import ID: pcg_id_or_name:dns_map_id_or_name
	// Supports: pcg_id:dns_map_id, pcg_name:dns_map_name, pcg_id:dns_map_name, pcg_name:dns_map_id
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("DNS map import ID is required")
	}

	parts := strings.Split(importID, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format. Expected pcg_id_or_name:dns_map_id_or_name, got: %s", importID)
	}

	pcgPart := parts[0]
	dnsMapPart := parts[1]

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

	// Resolve DNS map: try by ID first (and verify it belongs to this PCG), then by name via list
	var dnsMap *models.V1VsphereDNSMapping
	dnsMap, err = c.GetVsphereDNSMap(dnsMapPart)
	if err == nil && dnsMap != nil {
		// Verify DNS map belongs to the resolved PCG (for case pcg_name:dns_map_id)
		if dnsMap.Spec != nil && dnsMap.Spec.PrivateGatewayUID != nil && *dnsMap.Spec.PrivateGatewayUID != pcgUID {
			return nil, fmt.Errorf("DNS map does not belong to the specified private cloud gateway")
		}
	} else {
		// Resolve by name: list DNS maps for this PCG and find by Spec.DNSName or Metadata.Name
		list, listErr := c.GetVsphereDNSMappingsByPCGId(pcgUID)
		if listErr != nil {
			return nil, fmt.Errorf("unable to list DNS maps for private cloud gateway: %w", listErr)
		}
		var matches []*models.V1VsphereDNSMapping
		for _, item := range list.Items {
			if item.Spec != nil && item.Spec.DNSName != nil && *item.Spec.DNSName == dnsMapPart {
				matches = append(matches, item)
			} else if item.Metadata != nil && item.Metadata.Name == dnsMapPart {
				matches = append(matches, item)
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("DNS map not found for name or id: %s", dnsMapPart)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("multiple DNS maps match name '%s'; use DNS map ID for import", dnsMapPart)
		}
		dnsMap = matches[0]
	}

	// Set the required fields for the resource from the retrieved DNS map
	if err := d.Set("private_cloud_gateway_id", pcgUID); err != nil {
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

	// Set the ID to the DNS map UID
	if dnsMap.Metadata != nil && dnsMap.Metadata.UID != "" {
		d.SetId(dnsMap.Metadata.UID)
	} else {
		return nil, fmt.Errorf("resolved DNS map has no UID")
	}

	return c, nil
}
