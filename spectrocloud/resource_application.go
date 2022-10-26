package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
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
						"cluster_group_uid": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cluster_name": {
							Type:     schema.TypeString,
							Required: true,
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
										Required: true,
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

	application := toAppDeploymentClusterGroupEntity(d)

	diagnostics, isError := waitForClusterCreation(ctx, d, "" /*clusterUid*/, diags, c)
	if isError {
		return diagnostics
	}

	uid, err := c.CreateApplication(application)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	diagnostics, isError = waitForApplicationCreation(ctx, d, "" /*cluster.Metadata.UID*/, uid, diags, c)
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
		diagnostics, isError := waitForAddonDeploymentUpdate(ctx, d, cluster.Metadata.UID, addonDeployment.Profiles[0].UID, diags, c)
		if isError {
			return diagnostics
		}

		resourceApplicationRead(ctx, d, m)

		return diags
	}

	return diags
}

func toAppDeploymentClusterGroupEntity(d *schema.ResourceData) *models.V1AppDeploymentClusterGroupEntity {
	return &models.V1AppDeploymentClusterGroupEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name:   d.Get("name").(string),
			Labels: toTags(d),
		},
		Spec: toAppDeploymentClusterGroupSpec(d),
	}
}

func toAppDeploymentClusterGroupSpec(d *schema.ResourceData) *models.V1AppDeploymentClusterGroupSpec {
	return &models.V1AppDeploymentClusterGroupSpec{
		Config:  toV1AppDeploymentClusterGroupConfigEntity(d),
		Profile: toV1AppDeploymentProfileEntity(d),
	}
}

func toV1AppDeploymentClusterGroupConfigEntity(d *schema.ResourceData) *models.V1AppDeploymentClusterGroupConfigEntity {
	return &models.V1AppDeploymentClusterGroupConfigEntity{
		TargetSpec: toAppDeploymentClusterGroupTargetSpec(d),
	}
}

func toAppDeploymentClusterGroupTargetSpec(d *schema.ResourceData) *models.V1AppDeploymentClusterGroupTargetSpec {
	configList := d.Get("config")
	config := configList.([]interface{})[0].(map[string]interface{})

	return &models.V1AppDeploymentClusterGroupTargetSpec{
		ClusterGroupUID: ptr.StringPtr(config["cluster_group_uid"].(string)),
		ClusterLimits:   toAppDeploymentTargetClusterLimits(d),
		ClusterName:     ptr.StringPtr(config["cluster_name"].(string)),
	}
}

func toAppDeploymentTargetClusterLimits(d *schema.ResourceData) *models.V1AppDeploymentTargetClusterLimits {
	configList := d.Get("config")
	config := configList.([]interface{})[0].(map[string]interface{})
	limits := config["limits"].([]interface{})[0].(map[string]interface{})

	return &models.V1AppDeploymentTargetClusterLimits{
		CPU:       int32(limits["cpu"].(int)),
		MemoryMiB: int32(limits["memory"].(int)),
	}
}

func toV1AppDeploymentProfileEntity(d *schema.ResourceData) *models.V1AppDeploymentProfileEntity {
	return &models.V1AppDeploymentProfileEntity{
		AppProfileUID: ptr.StringPtr(d.Get("application_profile_uid").(string)),
	}
}
