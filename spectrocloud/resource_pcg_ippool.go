package spectrocloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
)

func resourcePrivateCloudGatewayIpPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIpPoolCreate,
		ReadContext:   resourceIpPoolRead,
		UpdateContext: resourceIpPoolUpdate,
		DeleteContext: resourceIpPoolDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"private_cloud_gateway_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"network_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateNetworkType,
			},
			"ip_start_range": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ip_end_range": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subnet_cidr": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"prefix": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Required: true,
			},
			"nameserver_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"nameserver_search_suffix": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"restrict_to_single_cluster": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceIpPoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := GetResourceLevelV1Client(m, "")
	var diags diag.Diagnostics
	pcgUID := d.Get("private_cloud_gateway_id").(string)

	pool := toIpPool(d)

	uid, err := c.CreateIpPool(pcgUID, pool)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func resourceIpPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := GetResourceLevelV1Client(m, "")
	var diags diag.Diagnostics

	pcgUID := d.Get("private_cloud_gateway_id").(string)

	pool, err := c.GetIpPool(pcgUID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	} else if pool == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", pool.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("gateway", pool.Spec.Pool.Gateway); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("prefix", pool.Spec.Pool.Prefix); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("restrict_to_single_cluster", pool.Spec.RestrictToSingleCluster); err != nil {
		return diag.FromErr(err)
	}

	if len(pool.Spec.Pool.Subnet) > 0 {
		if err := d.Set("subnet_cidr", pool.Spec.Pool.Subnet); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("ip_start_range", pool.Spec.Pool.Start); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("ip_end_range", pool.Spec.Pool.End); err != nil {
			return diag.FromErr(err)
		}
	}

	if pool.Spec.Pool.Nameserver != nil && len(pool.Spec.Pool.Nameserver.Addresses) > 0 {
		if err := d.Set("nameserver_addresses", pool.Spec.Pool.Nameserver.Addresses); err != nil {
			return diag.FromErr(err)
		}
	} else if pool.Spec.Pool.Nameserver != nil && len(pool.Spec.Pool.Nameserver.Search) > 0 {
		if err := d.Set("nameserver_search_suffix", pool.Spec.Pool.Nameserver.Search); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceIpPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := GetResourceLevelV1Client(m, "")
	var diags diag.Diagnostics

	pcgUID := d.Get("private_cloud_gateway_id").(string)

	pool := toIpPool(d)

	err := c.UpdateIpPool(pcgUID, d.Id(), pool)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceIpPoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := GetResourceLevelV1Client(m, "")
	var diags diag.Diagnostics

	pcgUID := d.Get("private_cloud_gateway_id").(string)

	err := c.DeleteIpPool(pcgUID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toIpPool(d *schema.ResourceData) *models.V1IPPoolInputEntity {
	pool := &models.V1Pool{
		Gateway:    d.Get("gateway").(string),
		Nameserver: &models.V1Nameserver{},
		Prefix:     int32(d.Get("prefix").(int)),
	}

	if d.Get("network_type").(string) == "range" {
		pool.Start = d.Get("ip_start_range").(string)
		pool.End = d.Get("ip_end_range").(string)
	} else {
		pool.Subnet = d.Get("subnet_cidr").(string)
	}

	if d.Get("nameserver_addresses") != nil {
		addresses := make([]string, 0)
		for _, az := range d.Get("nameserver_addresses").(*schema.Set).List() {
			addresses = append(addresses, az.(string))
		}
		pool.Nameserver.Addresses = addresses
	}

	if d.Get("nameserver_search_suffix") != nil {
		searchArr := make([]string, 0)
		for _, az := range d.Get("nameserver_search_suffix").(*schema.Set).List() {
			searchArr = append(searchArr, az.(string))
		}
		pool.Nameserver.Search = searchArr
	}

	return &models.V1IPPoolInputEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1IPPoolInputEntitySpec{
			Pool:                    pool,
			RestrictToSingleCluster: d.Get("restrict_to_single_cluster").(bool),
		},
	}
}

func validateNetworkType(data interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	networkType := data.(string)
	for _, nwType := range []string{"range", "subnet"} {
		if nwType == networkType {
			return diags
		}
	}
	return diag.FromErr(fmt.Errorf("network type '%s' is invalid. valid network types are 'range' and 'subnet'", networkType))
}
