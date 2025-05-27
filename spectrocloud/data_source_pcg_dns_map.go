package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourcePrivateCloudGatewayDNSMap() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDNSMapRead,

		Schema: map[string]*schema.Schema{
			"search_domain_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The domain name used for DNS search queries within the private cloud.",
			},
			"network": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The network to which the private cloud gateway is mapped.",
			},
			"data_center": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The data center in which the private cloud resides.",
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the private cloud gateway.",
			},
		},
	}
}

func dataSourceDNSMapRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := getV1ClientWithResourceContext(m, "tenant")
	pcgUID := d.Get("private_cloud_gateway_id").(string)
	name := d.Get("search_domain_name").(string)
	network := d.Get("network").(string)

	DNSMappings, err := c.GetVsphereDNSMappingsByPCGId(pcgUID)
	if err != nil {
		return handleReadError(d, err, diags)
	}
	if len(DNSMappings.Items) == 0 {
		err := fmt.Errorf("ResourceNotFound: No DNS Mapping identified in private_cloud_gateway_id - `%s`", name)
		return handleReadError(d, err, diags)
	}
	matchDNSMap := &models.V1VsphereDNSMappings{}
	for _, dnsMap := range DNSMappings.Items {
		if name == *dnsMap.Spec.DNSName {
			if network != "" {
				if network == *dnsMap.Spec.Network {
					matchDNSMap.Items = append(matchDNSMap.Items, dnsMap)
				}
			} else {
				matchDNSMap.Items = append(matchDNSMap.Items, dnsMap)
			}
		}
	}
	if len(matchDNSMap.Items) == 0 {
		err := fmt.Errorf("error: No DNS Map identified for name `%s` and network `%s`. Kindly re-try with up valid `name` and `network`", name, network)
		return diag.FromErr(err)
	} else if len(matchDNSMap.Items) == 1 {
		err := setBackDNSMap(matchDNSMap.Items[0], d)
		if err != nil {
			return diag.FromErr(err)
		}
	} else if len(matchDNSMap.Items) > 1 {
		if network == "" {
			err := fmt.Errorf("error: more than 1 DNS Mapping identified for name `%s`. Kindly try setting up `network` and try again", name)
			return diag.FromErr(err)
		} else {
			for _, dns := range matchDNSMap.Items {
				if network == *dns.Spec.Network {
					err = setBackDNSMap(dns, d)
					if err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}
	}
	return diags
}

func setBackDNSMap(dnsMap *models.V1VsphereDNSMapping, d *schema.ResourceData) error {
	d.SetId(dnsMap.Metadata.UID)
	err := d.Set("private_cloud_gateway_id", dnsMap.Spec.PrivateGatewayUID)
	if err != nil {
		return err
	}
	err = d.Set("search_domain_name", dnsMap.Spec.DNSName)
	if err != nil {
		return err
	}
	err = d.Set("network", dnsMap.Spec.Network)
	if err != nil {
		return err
	}
	err = d.Set("data_center", dnsMap.Spec.Datacenter)
	if err != nil {
		return err
	}
	return nil
}
