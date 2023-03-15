package spectrocloud

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
)

func resourceVirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualMachineCreate,
		ReadContext:   resourceVirtualMachineRead,
		UpdateContext: resourceVirtualMachineUpdate,
		DeleteContext: resourceVirtualMachineDelete,

		Schema: map[string]*schema.Schema{
			"cluster_uid": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cluster UID to which the virtual machine belongs to.",
			},
			"clone_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of the virtual machine to be cloned.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the virtual machine.",
			},
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The namespace of the virtual machine.",
			},
			"labels": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The labels of the virtual machine.",
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The annotations of the virtual machine.",
			},
			"vm_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "start",
				ValidateFunc: validation.StringInSlice([]string{"start", "stop", "restart", "pause", "resume", "migrate"}, false),
				Description:  "The action to be performed on the virtual machine. Valid values are: `start`, `stop`, `restart`, `pause`, `resume`, `migrate`. Default value is `start`.",
			},
			"vm_state": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The state of the virtual machine. Some of possible values are: `running`, `stopped`, `paused`, `migrating`, `error`, `unknown`.",
			},
			"cpu_cores": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The number of CPU cores to be allocated to the virtual machine. Default value is `1`.",
			},
			"run_on_launch": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If set to `true`, the virtual machine will be started when the cluster is launched. Default value is `true`.",
			},
			"memory": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "2G",
				Description: "The amount of memory to be allocated to the virtual machine. Default value is `2G`.",
			},
			"image_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"volume_spec"},
				Description:   "The URL of the image(template) to be used for the virtual machine.",
			},
			"cloud_init_user_data": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n",
				ConflictsWith: []string{"volume_spec"},
				Description:   "The cloud-init user data to be used for the virtual machine. Default value is `#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n`.",
			},
			"devices": schemas.VMDeviceSchema(),
			"volume_spec": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"volume": schemas.VMVolumeSchema(),
					},
				},
				Description: "Volume specification for the virtual machine.",
			},
			"network_spec": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the network to be attached to the virtual machine.",
									},
								},
							},
							Description: "Network specification for the virtual machine.",
						},
					},
				},
			},
		},
	}
}

func resourceVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	clusterUid := d.Get("cluster_uid").(string)
	cluster, err := c.GetCluster(clusterUid)
	if err != nil {
		return diag.FromErr(err)
	}
	var vm *models.V1ClusterVirtualMachine
	if cloneFromVM, ok := d.GetOk("clone_name"); ok && cloneFromVM != "" {
		name := d.Get("name").(string)
		nameSpace := d.Get("namespace").(string)
		err := c.CloneVirtualMachine(clusterUid, cloneFromVM.(string), name, nameSpace)
		if err != nil {
			return diag.FromErr(err)
		}
		vm, err = c.GetVirtualMachine(clusterUid, name, nameSpace)
		if d.Get("run_on_launch").(bool) {
			diags = resourceVirtualMachineActions(c, ctx, d, "start", clusterUid, name, nameSpace)
			if diags.HasError() {
				return diags
			}
		}
		d.SetId(name)
	} else {
		virtualMachine, err := toVirtualMachineCreateRequest(d)
		if err != nil {
			return diag.FromErr(err)
		}
		vm, err = c.CreateVirtualMachine(cluster.Metadata.UID, virtualMachine)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.Get("run_on_launch").(bool) {
			diags, _ = waitForVirtualMachineToTargetState(ctx, d, d.Get("cluster_uid").(string), d.Get("name").(string), d.Get("namespace").(string), diags, c, "create", "Running")
			if diags.HasError() {
				return diags
			}
		}
		d.SetId(vm.Metadata.Name)
	}

	resourceVirtualMachineRead(ctx, d, m)
	return diags
}

func resourceVirtualMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	// Read the virtual machine name and namespace from the resource data
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)

	// Call the client's method to retrieve the virtual machine details
	vm, err := c.GetVirtualMachine(d.Get("cluster_uid").(string), name, namespace)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update the resource data with the retrieved virtual machine metadata details
	d.SetId(vm.Metadata.Name)
	d.Set("name", vm.Metadata.Name)
	d.Set("namespace", vm.Metadata.Namespace)

	if _, ok := d.GetOk("labels"); ok {
		d.Set("labels", flattenVMLabels(vm.Metadata.Labels))
	}

	if _, ok := d.GetOk("annotations"); ok {
		d.Set("annotations", flattenVMAnnotations(vm.Metadata.Annotations, d))
	}

	domain := vm.Spec.Template.Spec.Domain
	volume := vm.Spec.Template.Spec.Volumes
	if domain.CPU != nil {
		d.Set("cpu_cores", domain.CPU.Cores)
	}
	if domain.Resources != nil {
		if domain.Resources.Requests != nil {
			if memory := domain.Resources.Requests.(map[string]interface{})["memory"]; memory != nil && memory != "" {
				d.Set("memory", memory.(string))
			}
		}
	}
	if _, imgOk := d.GetOk("image_url"); imgOk {
		if _, volOk := d.GetOk("volume"); !volOk {
			for _, v := range volume {
				if v.ContainerDisk != nil {
					d.Set("image_url", v.ContainerDisk.Image)
				}
			}
		}
	}
	d.Set("vm_state", vm.Status.PrintableStatus)
	// setting back network
	if _, ok := d.GetOk("network_spec"); ok && vm.Spec.Template.Spec.Networks != nil {
		d.Set("network", flattenVMNetwork(vm.Spec.Template.Spec.Networks))
	}

	// setting back volume
	if _, ok := d.GetOk("volume_spec"); ok && vm.Spec.Template.Spec.Volumes != nil {
		d.Set("volume", flattenVMVolumes(vm.Spec.Template.Spec.Volumes))
	}
	// setting back devices
	if _, ok := d.GetOk("devices"); ok && domain.Devices != nil {
		d.Set("devices", flattenVMDevices(d, domain.Devices))
	}

	return diags
}

func resourceVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	clusterUid := d.Get("cluster_uid").(string)
	vmName := d.Get("name").(string)
	vmNamespace := d.Get("namespace").(string)
	vm, err := c.GetVirtualMachine(clusterUid, vmName, vmNamespace)
	if err != nil {
		return diag.FromErr(err)
	}

	cluster, err := c.GetCluster(clusterUid)
	if err != nil {
		return diag.FromErr(err)
	}

	needUpdate, needRestart, updatedVmModel, err := toVirtualMachineUpdateRequest(d, vm)
	if err != nil {
		return diag.FromErr(err)
	}

	if needUpdate {
		_, err = c.UpdateVirtualMachine(cluster, vmName, updatedVmModel)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if needRestart {
		stateToChange := "restart"
		resourceVirtualMachineActions(c, ctx, d, stateToChange, clusterUid, vmName, vmNamespace)
	}
	if _, ok := d.GetOk("vm_action"); ok && d.HasChange("vm_action") {
		stateToChange := d.Get("vm_action").(string)
		resourceVirtualMachineActions(c, ctx, d, stateToChange, clusterUid, vmName, vmNamespace)
	}

	return resourceVirtualMachineRead(ctx, d, m)
}

func resourceVirtualMachineActions(c *client.V1Client, ctx context.Context, d *schema.ResourceData, stateToChange string, clusterUid string, vmName string, vmNamespace string) diag.Diagnostics {
	var diags diag.Diagnostics
	// need to add validation status and allowed actions
	// Stopped  - start
	// Paused - restart, resume
	// Running - stop ,restart,pause, migrate
	switch strings.ToLower(stateToChange) {
	//"start", "stop", "restart", "pause", "resume", "migrate"
	case "start":
		err := c.StartVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
		break
	case "stop":
		err := c.StopVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Stopped")
		if diags.HasError() {
			return diags
		}
		break
	case "restart":
		err := c.RestartVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
		break
	case "pause":
		err := c.PauseVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Paused")
		if diags.HasError() {
			return diags
		}
		break
	case "resume":
		err := c.ResumeVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
		break
	case "migrate":
		err := c.MigrateVirtualMachineNodeToNode(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
		break
	}
	vm, err := c.GetVirtualMachine(clusterUid, vmName, vmNamespace)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("vm_state", vm.Status.PrintableStatus)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	err := c.DeleteVirtualMachine(d.Get("cluster_uid").(string), d.Get("name").(string), d.Get("namespace").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
