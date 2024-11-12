package spectrocloud

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func resourcePrivateCloudGatewayDNSMap() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePCGDNSMapCreate,
		ReadContext:   resourcePCGDNSMapRead,
		UpdateContext: resourcePCGDNSMapUpdate,
		DeleteContext: resourcePCGDNSMapDelete,
		Description:   "This resource allows for the management of DNS mappings for private cloud gateways. This helps ensure proper DNS resolution for resources within the private cloud environment.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The ID of the Private Cloud Gateway.",
			},
			"search_domain_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The domain name used for DNS search queries within the private cloud.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`),
					"must be a valid domain name",
				),
			},
			"data_center": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The data center in which the private cloud resides.",
			},
			"network": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The network to which the private cloud gateway is mapped.",
			},
		},
	}
}

func toDNSMap(d *schema.ResourceData) *models.V1VsphereDNSMapping {
	return &models.V1VsphereDNSMapping{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("search_domain_name").(string),
		},
		Spec: &models.V1VsphereDNSMappingSpec{
			Datacenter:        ptr.To(d.Get("data_center").(string)),
			DNSName:           ptr.To(d.Get("search_domain_name").(string)),
			Network:           ptr.To(d.Get("network").(string)),
			PrivateGatewayUID: ptr.To(d.Get("private_cloud_gateway_id").(string)),
			// UI doesn't send network_url may need to enable in the future.
			// NetworkURL:        "",
		},
	}
}

func flattenDNSMap(dnsMap *models.V1VsphereDNSMapping, d *schema.ResourceData) error {
	if dnsMap != nil {
		if *dnsMap.Spec.DNSName != "" {
			err := d.Set("search_domain_name", *dnsMap.Spec.DNSName)
			if err != nil {
				return err
			}
		}
		if *dnsMap.Spec.Datacenter != "" {
			err := d.Set("data_center", *dnsMap.Spec.Datacenter)
			if err != nil {
				return err
			}
		}
		if *dnsMap.Spec.Network != "" {
			err := d.Set("network", *dnsMap.Spec.Network)
			if err != nil {
				return err
			}
		}
		if *dnsMap.Spec.PrivateGatewayUID != "" {
			err := d.Set("private_cloud_gateway_id", *dnsMap.Spec.PrivateGatewayUID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func resourcePCGDNSMapCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	dnsMap := toDNSMap(d)
	uid, err := c.CreateVsphereDNSMap(dnsMap)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func resourcePCGDNSMapUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	dnsMap := toDNSMap(d)
	err := c.UpdateVsphereDNSMap(d.Id(), dnsMap)

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourcePCGDNSMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	dnsMap, err := c.GetVsphereDNSMap(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenDNSMap(dnsMap, d)
	if err != nil {
		return nil
	}
	return diags
}

func resourcePCGDNSMapDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	err := c.DeleteVsphereDNSMap(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
