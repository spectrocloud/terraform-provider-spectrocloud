package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"time"
)

func resourceResourceLimit() *schema.Resource {
	return &schema.Resource{
		CreateContext: nil,
		ReadContext:   nil,
		UpdateContext: nil,
		DeleteContext: nil,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Description:   "",
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"alert": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 10000),
				Description:  "The number of alert limit. Must be between 1 and 10000.",
			},
			"api_keys": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"application_deployment": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"application_profile": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"certificate": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"cloud_account": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"cluster_group": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"cluster_profile": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"appliance": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"appliance_token": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"filter": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"location": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"macro": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"private_gateway": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"project": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"registry": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"role": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"cluster": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"ssh_key": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"team": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"user": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
			"workspace": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(1, 10000),
			},
		},
	}
}
