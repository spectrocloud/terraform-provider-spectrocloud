package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationCreate,
		ReadContext:   resourceApplicationRead,
		UpdateContext: resourceApplicationUpdate,
		DeleteContext: resourceApplicationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"application_profile_uid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_uid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cluster_group_uid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cluster_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"limits": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  "spectro",
									},
									"memory": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"storage": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	/*clusterUid := d.Get("cluster_group_uid").(string)

	cluster, err := c.GetCluster(clusterUid)
	if err != nil && cluster == nil {
		return diag.FromErr(errors.New(fmt.Sprintf("Cluster not found: %s", clusterUid)))
	}*/
	val_error := errors.New("config block should have either 'cluster_uid' or 'cluster_group_uid' attributes specified.")

	var uid string
	var err error
	var config map[string]interface{}
	var cluster_uid interface{}
	configList := d.Get("config")
	if configList.([]interface{})[0] != nil {
		config = configList.([]interface{})[0].(map[string]interface{})
		cluster_uid = config["cluster_uid"]
	} else {
		return diag.FromErr(val_error)
	}

	if cluster_uid == nil {
		if config["cluster_group_uid"] == nil {
			return diag.FromErr(val_error)
		}
		application := toAppDeploymentClusterGroupEntity(d)

		/*diagnostics, isError := waitForClusterCreation(ctx, d, clusterUid, diags, c)
		if isError {
			return diagnostics
		}*/

		uid, err = c.CreateApplicationWithNewSandboxCluster(application)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		application := toAppDeploymentNestedClusterEntity(d)

		/*diagnostics, isError := waitForClusterCreation(ctx, d, clusterUid, diags, c)
		if isError {
			return diagnostics
		}*/

		uid, err = c.CreateApplicationWithExistingSandboxCluster(application)
		if err != nil {
			return diag.FromErr(err)
		}

	}

	d.SetId(uid)

	diagnostics, isError := waitForApplicationCreation(ctx, d, diags, c)
	if isError {
		return diagnostics
	}

	resourceApplicationRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceApplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	/*c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Get("cluster_uid").(string)
	cluster, err := c.GetCluster(uid)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}*/
	var diags diag.Diagnostics

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("cluster_uid", "cluster_profile") {
		c := m.(*client.V1Client)

		clusterUid := d.Get("cluster_uid").(string)

		cluster, err := c.GetCluster(clusterUid)
		if err != nil && cluster == nil {
			return diag.FromErr(errors.New(fmt.Sprintf("Cluster not found: %s", clusterUid)))
		}

		addonDeployment := toAddonDeployment(c, d)

		newProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
		err = c.UpdateAddonDeployment(cluster, addonDeployment, newProfile)
		if err != nil {
			return diag.FromErr(err)
		}

		clusterProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(getAddonDeploymentId(clusterUid, clusterProfile))
		diagnostics, isError := waitForApplicationUpdate(ctx, d, diags, c)
		if isError {
			return diagnostics
		}

		resourceApplicationRead(ctx, d, m)

		return diags
	}

	return diags
}
