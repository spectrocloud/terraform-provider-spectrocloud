package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the application being created.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Set:         schema.HashString,
				Description: "A set of tags to associate with the application for easier identification and categorization.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"application_profile_uid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier (UID) of the application profile to use for this application.",
			},
			"config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The configuration block for specifying cluster and resource limits for the application.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_uid": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The unique identifier (UID) of the target cluster. Either `cluster_uid` or `cluster_group_uid` can be provided.",
						},
						"cluster_group_uid": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The unique identifier (UID) of the cluster group. Either `cluster_uid` or `cluster_group_uid` can be provided.",
						},
						"cluster_context": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The context for the cluster,  Either `tenant` or `project` can be provided.",
						},
						"cluster_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An optional name for the target cluster.",
						},
						"limits": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Optional resource limits for the application, including CPU, memory, and storage constraints.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "The CPU allocation for the application, specified in integer values.",
									},
									"memory": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "The memory allocation for the application, specified in megabytes.",
									},
									"storage": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "The storage allocation for the application, specified in gigabytes.",
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
	resourceContext := ""
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	/*clusterUid := d.Get("cluster_group_uid").(string)

	cluster, err := c.GetCluster(clusterUid)
	if err != nil && cluster == nil {
		return diag.FromErr(errors.New(fmt.Sprintf("Cluster not found: %s", clusterUid)))
	}*/
	val_error := errors.New("config block should have either 'cluster_uid' or 'cluster_group_uid' attributes specified")

	var uid string
	var err error
	var config map[string]interface{}
	var cluster_uid interface{}
	configList := d.Get("config")
	if configList.([]interface{})[0] != nil {

		config = configList.([]interface{})[0].(map[string]interface{})
		cluster_uid = config["cluster_uid"]
		resourceContext = config["cluster_context"].(string)

	} else {
		return diag.FromErr(val_error)
	}
	c := getV1ClientWithResourceContext(m, resourceContext)
	if cluster_uid == "" {
		if config["cluster_group_uid"] == "" {
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
		application := toAppDeploymentVirtualClusterEntity(d)

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

	if d.HasChanges("config.0.cluster_uid", "config.0.cluster_profile") {
		configList := d.Get("config")
		c := getV1ClientWithResourceContext(m, "")
		if configList.([]interface{})[0] != nil {
			config := configList.([]interface{})[0].(map[string]interface{})
			resourceContext := config["cluster_context"].(string)
			c = getV1ClientWithResourceContext(m, resourceContext)
		}

		clusterUid := d.Get("cluster_uid").(string)
		cluster, err := c.GetCluster(clusterUid)
		if err != nil && cluster == nil {
			return diag.FromErr(fmt.Errorf("cluster not found: %s", clusterUid))
		}

		addonDeployment, err := toAddonDeployment(c, d)
		if err != nil {
			return diag.FromErr(err)
		}

		newProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
		if err != nil {
			return diag.FromErr(err)
		}
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
