package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"strings"
)

func resourceVirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualMachineCreate,
		ReadContext:   resourceVirtualMachineRead,
		UpdateContext: resourceVirtualMachineUpdate,
		DeleteContext: resourceVirtualMachineDelete,

		Schema: map[string]*schema.Schema{
			"cluster_uid": {
				Type:     schema.TypeString,
				Required: true,
				//ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"labels": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
				//ForceNew: true,
			},
			"vm_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Running",
				ValidateFunc: validation.StringInSlice([]string{"start", "stop", "restart", "pause", "clone", "resume", "migrate", ""}, false),
			},
			"vm_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cpu_cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"run_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"memory": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2G",
			},
			"image_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"volume"},
			},
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cloud_init_user_data": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n",
				ConflictsWith: []string{"volume"},
			},
			"devices": schemas.VMDeviceSchema(),
			"volume":  schemas.VMVolumeSchema(),
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
	if err != nil && cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster not found: %s", clusterUid))
	}
	virtualMachine, err := toVirtualMachineCreateRequest(d)
	if err != nil {
		return diag.FromErr(err)
	}
	vm, err := c.CreateVirtualMachine(cluster.Metadata.UID, virtualMachine)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vm.Metadata.Name)
	if d.Get("run_on_launch").(bool) {
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, d.Get("cluster_uid").(string), d.Get("name").(string), d.Get("namespace").(string), diags, c, "create", "Running")
		if diags.HasError() {
			return diags
		}
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
	if _, ok := d.GetOk("network"); ok && vm.Spec.Template.Spec.Networks != nil {
		d.Set("network", flattenVMNetwork(vm.Spec.Template.Spec.Networks))
	}

	// setting back volume
	if _, ok := d.GetOk("volume"); ok && vm.Spec.Template.Spec.Volumes != nil {
		d.Set("volume", flattenVMVolumes(vm.Spec.Template.Spec.Volumes))
	}
	// setting back devices
	d.Set("devices", flattenVMDevices(d, domain.Devices))

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

	needUpdate, updatedVmModel, err := toVirtualMachineUpdateRequest(d, vm)
	if err != nil {
		return diag.FromErr(err)
	}

	if needUpdate {
		_, err = c.UpdateVirtualMachine(cluster, vmName, updatedVmModel)
		if err != nil {
			return diag.FromErr(err)
		}
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
	// Stopped  - start,clone
	// Paused - restart, resume, clone
	// Running - stop ,restart,pause,clone, migrate
	switch strings.ToLower(stateToChange) {
	//"start", "stop", "restart", "pause", "clone", "resume", "migrate"
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
	case "clone":
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
