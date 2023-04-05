package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/convert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/virtualmachine"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

func resourceKubevirtVirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubevirtVirtualMachineCreate,
		ReadContext:   resourceKubevirtVirtualMachineRead,
		UpdateContext: resourceVirtualMachineUpdate,
		DeleteContext: resourceKubevirtVirtualMachineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: virtualmachine.VirtualMachineFields(),
	}
}
func resourceKubevirtVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	clusterUid := d.Get("cluster_uid").(string)
	cluster, err := c.GetCluster(clusterUid)
	if err != nil {
		return diag.FromErr(err)
	}
	// if cluster is nil(deleted or not found), return error
	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster not found for uid %s", clusterUid))
	}
	virtualMachineToCreate, err := virtualmachine.FromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	hapiVM, err := convert.ToHapiVm(virtualMachineToCreate)
	if _, ok := d.GetOk("run_on_launch"); ok {
		if !d.Get("run_on_launch").(bool) {
			hapiVM.Spec.RunStrategy = "Manual"
		} else {
			hapiVM.Spec.Running = d.Get("run_on_launch").(bool)
		}
	}

	if err != nil {
		return diag.FromErr(err)
	}
	if cloneFromVM, ok := d.GetOk("base_vm_name"); ok && cloneFromVM != "" {
		// Handling clone case
		err = c.CloneVirtualMachine(clusterUid, cloneFromVM.(string), hapiVM.Metadata.Name, hapiVM.Metadata.Namespace)
		if err != nil {
			return diag.FromErr(err)
		}
		vm, err := c.GetVirtualMachine(clusterUid, hapiVM.Metadata.Namespace, hapiVM.Metadata.Name)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(utils.BuildId(clusterUid, vm.Metadata))
		// apply the rest of configuration after clone to override it.
		hapiVM.Metadata.ResourceVersion = vm.Metadata.ResourceVersion // set resource version to avoid conflict
		/*		//	// TODO: There is issue in Ally side, team asked as to explicitly make deletion-time to nil before put operation, after fix will remove.
				hapiVM.Spec.Template.Metadata.DeletionTimestamp = nil
				hapiVM.Metadata.DeletionTimestamp = nil
				hapiVM.Spec.Template.Metadata.CreationTimestamp = ""
				hapiVM.Metadata.CreationTimestamp = ""*/
		_, err = c.UpdateVirtualMachine(cluster, hapiVM.Metadata.Name, hapiVM)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		vm, err := c.CreateVirtualMachine(cluster.Metadata.UID, hapiVM)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.Get("run_on_launch").(bool) {
			diags, _ = waitForVirtualMachineToTargetState(ctx, d, cluster.Metadata.UID, hapiVM.Metadata.Name, hapiVM.Metadata.Namespace, diags, c, "create", "Running")
			if diags.HasError() {
				return diags
			}
		}
		d.SetId(utils.BuildId(clusterUid, vm.Metadata))
	}

	resourceKubevirtVirtualMachineRead(ctx, d, m)
	return diags
}

/*func resourceKubevirtVirtualMachineCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := (meta).(*client.V1Client)

	vm, err := virtualmachine.FromResourceData(resourceData)
	if err != nil {
		return diag.FromErr(err)
	}

	hapiVM := convert.ToHapiVm(vm)
	log.Printf("[INFO] Creating new virtual machine: %#v", vm)
	if _, err := cli.CreateVirtualMachine(resourceData.Get("cluster_uid").(string), hapiVM); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Submitted new virtual machine: %#v", vm)
	if err := virtualmachine.ToResourceData(*vm, resourceData); err != nil {
		return diag.FromErr(err)
	}
	resourceData.SetId(utils.BuildId(vm.ObjectMeta))

	// Wait for virtual machine instance's status phase to be succeeded:
	name := vm.ObjectMeta.Name
	namespace := vm.ObjectMeta.Namespace

	stateConf := &resource.StateChangeConf{
		Pending: []string{"Creating"},
		Target:  []string{"Succeeded"},
		Timeout: resourceData.Timeout(schema.TimeoutCreate),
		Refresh: func() (interface{}, string, error) {
			var err error
			hapiVM, err = cli.GetVirtualMachine(resourceData.Get("cluster_uid").(string), namespace, name)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Printf("[DEBUG] virtual machine %s is not created yet", name)
					return vm, "Creating", nil
				}
				return vm, "", err
			}

			vm = convert.ToKubevirtVM(hapiVM)

			if vm == nil {
				return vm, "Error", fmt.Errorf("virtual machine %s is nil = probablly manuallly deleted.", name)
			}

			if vm.Status.Created == true && vm.Status.Ready == true {
				return vm, "Succeeded", nil
			}

			log.Printf("[DEBUG] virtual machine %s is being created", name)
			return vm, "Creating", nil
		},
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return diag.FromErr(fmt.Errorf("%s", err))
	}

	return resourceKubevirtVirtualMachineRead(ctx, resourceData, meta)
}*/

func resourceKubevirtVirtualMachineRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := (meta).(*client.V1Client)

	clusterUid, namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading virtual machine %s", name)

	hapiVM, err := cli.GetVirtualMachine(clusterUid, namespace, name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return diag.FromErr(err)
	}
	vm, err := convert.ToKubevirtVM(hapiVM)
	if err != nil {
		return diag.FromErr(err)
	}
	if vm == nil {
		return nil
	}
	log.Printf("[INFO] Received virtual machine: %#v", vm)

	err = virtualmachine.ToResourceData(*vm, resourceData)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
func resourceVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	/*func resourceKubevirtVirtualMachineUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {

		ops := virtualmachine.AppendPatchOps("", "", resourceData, make([]patch.PatchOperation, 0, 0))
		data, err := ops.MarshalJSON()
		if err != nil {
			return diag.FromErr(fmt.Errorf("Failed to marshal update operations: %s", err))
		}

		log.Printf("[INFO] Updating virtual machine: %s", ops)
		out := &kubevirtapiv1.VirtualMachine{}
		//	have (string, string, *"kubevirt.io/api/core/v1".VirtualMachine, []byte)
		//	want (*models.V1SpectroCluster, string, *models.V1ClusterVirtualMachine)
		// if _, err := cli.UpdateVirtualMachine(&models.V1SpectroCluster{}, namespace, name, out, data); err != nil {
		if _, err := cli.UpdateVirtualMachine(&models.V1SpectroCluster{}, namespace, name, &models.V1ClusterVirtualMachine{}, data); err != nil {
			return diag.FromErr(err)
		}

		log.Printf("[INFO] Submitted updated virtual machine: %#v", out)

		return resourceKubevirtVirtualMachineRead(ctx, resourceData, meta)
	}*/
	c := m.(*client.V1Client)
	clusterUid, vmNamespace, vmName, err := utils.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.GetVirtualMachine(clusterUid, vmNamespace, vmName)
	if err != nil {
		return diag.FromErr(err)
	}

	// prepare new vm data
	vm, err := virtualmachine.FromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}
	hapiVM, err := convert.ToHapiVm(vm)
	if err != nil {
		return diag.FromErr(err)
	}

	// needed to get context for the cluster
	cluster, err := c.GetCluster(clusterUid)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("run_on_launch"); ok {
		if !d.Get("run_on_launch").(bool) {
			hapiVM.Spec.RunStrategy = "Manual"
		} else {
			hapiVM.Spec.Running = d.Get("run_on_launch").(bool)
		}
	}
	_, err = c.UpdateVirtualMachine(cluster, vmName, hapiVM)
	if err != nil {
		return diag.FromErr(err)
	}
	// Currently Restarts are handled via vm_actions through config later will remove below code
	//needUpdate, needRestart, _, err := toVirtualMachineUpdateRequest(d, currentVm)
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//if needUpdate {
	//	// TODO: There is issue in Ally side, team asked as to explicitly make deletion-time to nil before put operation, after fix will remove.
	//	//hapiVM.Spec.Template.Metadata.DeletionTimestamp = nil
	//	//hapiVM.Metadata.DeletionTimestamp = nil
	//	if _, ok := d.GetOk("run_on_launch"); ok {
	//		if !d.Get("run_on_launch").(bool) {
	//			hapiVM.Spec.RunStrategy = "Manual"
	//		} else {
	//			hapiVM.Spec.Running = d.Get("run_on_launch").(bool)
	//		}
	//	}
	//	_, err = c.UpdateVirtualMachine(cluster, vmName, hapiVM)
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//}
	//if needRestart {
	//	stateToChange := "restart"
	//	resourceVirtualMachineActions(c, ctx, d, stateToChange, clusterUid, vmName, vmNamespace)
	//}
	if _, ok := d.GetOk("vm_action"); ok && d.HasChange("vm_action") {
		stateToChange := d.Get("vm_action").(string)
		resourceVirtualMachineActions(c, ctx, d, stateToChange, clusterUid, vmName, vmNamespace)
	}

	return resourceKubevirtVirtualMachineRead(ctx, d, m)
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
	case "stop":
		err := c.StopVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Stopped")
		if diags.HasError() {
			return diags
		}
	case "restart":
		err := c.RestartVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
	case "pause":
		err := c.PauseVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Paused")
		if diags.HasError() {
			return diags
		}
	case "resume":
		err := c.ResumeVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
	case "migrate":
		_ = c.MigrateVirtualMachineNodeToNode(clusterUid, vmName, vmNamespace)
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
	}
	_, err := c.GetVirtualMachine(clusterUid, vmNamespace, vmName)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceKubevirtVirtualMachineDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterUid, namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cli := (meta).(*client.V1Client)

	log.Printf("[INFO] Deleting virtual machine: %#v", name)
	if err := cli.DeleteVirtualMachine(clusterUid, namespace, name); err != nil {
		return diag.FromErr(err)
	}

	// Wait for virtual machine instance to be removed:
	stateConf := &retry.StateChangeConf{
		Pending: []string{"Deleting"},
		Timeout: resourceData.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			vm, err := cli.GetVirtualMachine(resourceData.Get("cluster_uid").(string), namespace, name)
			if err != nil {
				if errors.IsNotFound(err) {
					return nil, "", nil
				}
				return vm, "", err
			}

			if vm == nil {
				return nil, "", nil
			}

			//log.Printf("[DEBUG] Virtual machine %s is being deleted", vm.GetName())
			return vm, "Deleting", nil
		},
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("%s", err))
	}

	log.Printf("[INFO] virtual machine %s deleted", name)

	resourceData.SetId("")
	return nil
}
