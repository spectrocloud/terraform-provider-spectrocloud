package spectrocloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterGroupCreate,
		ReadContext:   resourceClusterGroupRead,
		UpdateContext: resourceClusterGroupUpdate,
		DeleteContext: resourceClusterGroupDelete,

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
				ForceNew:    true,
				Description: "Name of the cluster group",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tenant",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description:  "Cluster group context can be 'project' or 'tenant'. Defaults to 'project'.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster group. Tags must be in the form of `key:value`.",
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_endpoint_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Ingress",
							ValidateFunc: validation.StringInSlice([]string{"", "Ingress", "LoadBalancer"}, false),
							Description:  "Host endpoint type can be 'Ingress' or 'LoadBalancer'. Defaults to 'Ingress'.",
						},
						"cpu_millicore": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "CPU limit in millicores.",
						},
						"memory_in_mb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Memory limit in MB.",
						},
						"storage_in_gb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Storage limit in GB.",
						},
						"oversubscription_percent": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Resource oversubscription percentage.",
						},
						"values": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},
			"clusters": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_uid": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The UID of the host cluster.",
						},
						"host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Host string in the cluster. I.e. *.dev or *.",
						},
					},
				},
			},
		},
	}
}

func resourceClusterGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cluster := toClusterGroup(d)

	uid, err := c.CreateClusterGroup(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceClusterGroupRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterGroupRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics
	//
	uid := d.Id()
	//
	cluster, err := c.GetClusterGroup(uid)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	return flattenClusterGroup(cluster, d)
}

func flattenClusterGroup(cluster *models.V1ClusterGroup, d *schema.ResourceData) diag.Diagnostics {

	return diag.Diagnostics{}
}

func resourceClusterGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	cloudConfigId := d.Get("cloud_config_id").(string)

	if d.HasChange("machine_pool") {
		oraw, nraw := d.GetChange("machine_pool")
		if oraw == nil {
			oraw = new(schema.Set)
		}
		if nraw == nil {
			nraw = new(schema.Set)
		}

		os := oraw.(*schema.Set)
		ns := nraw.(*schema.Set)

		osMap := make(map[string]interface{})
		for _, mp := range os.List() {
			machinePool := mp.(map[string]interface{})
			osMap[machinePool["name"].(string)] = machinePool
		}

		for _, mp := range ns.List() {
			machinePoolResource := mp.(map[string]interface{})
			name := machinePoolResource["name"].(string)
			hash := resourceMachinePoolVirtualHash(machinePoolResource)

			machinePool := toMachinePoolVirtual(machinePoolResource)

			var err error
			if oldMachinePool, ok := osMap[name]; !ok {
				log.Printf("Create machine pool %s", name)
				err = c.CreateMachinePoolVirtual(cloudConfigId, machinePool)
			} else if hash != resourceMachinePoolVirtualHash(oldMachinePool) {
				log.Printf("Change in machine pool %s", name)
				err = c.UpdateMachinePoolVirtual(cloudConfigId, machinePool)
			}
			if err != nil {
				return diag.FromErr(err)
			}

			// Processed (if exists)
			delete(osMap, name)
		}

		// Deleted old machine pools
		for _, mp := range osMap {
			machinePool := mp.(map[string]interface{})
			name := machinePool["name"].(string)
			log.Printf("Deleted machine pool %s", name)
			if err := c.DeleteMachinePoolVirtual(cloudConfigId, name); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diagnostics, done := updateCommonFields(d, c)
	if done {
		return diagnostics
	}

	resourceClusterGroupRead(ctx, d, m)

	return diags
}

func toClusterGroup(d *schema.ResourceData) *models.V1ClusterGroupEntity {
	clusterRefs := make([]*models.V1ClusterGroupClusterRef, 0)
	clusterRefObj, ok := d.GetOk("clusters")
	if ok {
		for i, _ := range clusterRefObj.([]interface{}) {
			resources := clusterRefObj.([]interface{})[i].(map[string]interface{})
			mp := toClusterRef(resources)
			clusterRefs = append(clusterRefs, mp)
		}
	}

	var clusterGroupLimitConfig *models.V1ClusterGroupLimitConfig
	resourcesObj, ok := d.GetOk("config")
	if ok {
		resources := resourcesObj.([]interface{})[0].(map[string]interface{})
		clusterGroupLimitConfig = toClusterGroupLimitConfig(resources)
	}

	ret := &models.V1ClusterGroupEntity{
		Metadata: &models.V1ObjectMeta{
			Name:   d.Get("name").(string),
			UID:    d.Id(),
			Labels: toTags(d),
		},
		Spec: &models.V1ClusterGroupSpec{
			Type:        "hostCluster",
			ClusterRefs: clusterRefs,
			ClustersConfig: &models.V1ClusterGroupClustersConfig{
				LimitConfig: clusterGroupLimitConfig,
			},
		},
	}

	return ret
}

func toClusterRef(resources map[string]interface{}) *models.V1ClusterGroupClusterRef {
	cluster_uid := resources["cluster_uid"].(string)

	ret := &models.V1ClusterGroupClusterRef{
		ClusterUID: cluster_uid,
	}

	return ret
}

func toClusterGroupLimitConfig(resources map[string]interface{}) *models.V1ClusterGroupLimitConfig {
	cpu_milli := resources["cpu_millicore"].(int)
	mem_in_mb := resources["memory_in_mb"].(int)
	storage_in_gb := resources["storage_in_gb"].(int)
	oversubscription := resources["oversubscription_percent"].(int)

	ret := &models.V1ClusterGroupLimitConfig{

		CPUMilliCore:     int32(cpu_milli),
		MemoryMiB:        int32(mem_in_mb),
		StorageGiB:       int32(storage_in_gb),
		OverSubscription: int32(oversubscription),
	}

	return ret
}

func resourceClusterGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	err := c.DeleteClusterGroup(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
