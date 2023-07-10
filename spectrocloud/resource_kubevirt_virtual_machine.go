package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
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
	ClusterContext := d.Get("cluster_context").(string)
	cluster, err := c.GetCluster(ClusterContext, clusterUid)
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
		err = c.CloneVirtualMachine(ClusterContext, clusterUid, cloneFromVM.(string), hapiVM.Metadata.Name, hapiVM.Metadata.Namespace)
		if err != nil {
			return diag.FromErr(err)
		}
		vm, err := c.GetVirtualMachine(ClusterContext, clusterUid, hapiVM.Metadata.Namespace, hapiVM.Metadata.Name)
		if err != nil {
			return diag.FromErr(err)
		}
		if vm == nil {
			return diag.FromErr(fmt.Errorf("virtual machine not found after clone operation %s, %s, %s", clusterUid, hapiVM.Metadata.Namespace, hapiVM.Metadata.Name))
		}
		d.SetId(utils.BuildId(ClusterContext, clusterUid, vm.Metadata))
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
		vm, err := c.CreateVirtualMachine(ClusterContext, cluster.Metadata.UID, hapiVM)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(utils.BuildId(ClusterContext, clusterUid, vm.Metadata))
	}
	if d.Get("run_on_launch").(bool) {
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, ClusterContext, cluster.Metadata.UID, hapiVM.Metadata.Name, hapiVM.Metadata.Namespace, diags, c, "create", "Running")
		if diags.HasError() {
			return diags
		}
	}

	resourceKubevirtVirtualMachineRead(ctx, d, m)
	return diags
}

func resourceKubevirtVirtualMachineRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := (meta).(*client.V1Client)

	scope, clusterUid, namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading virtual machine %s", name)

	hapiVM, err := cli.GetVirtualMachine(scope, clusterUid, namespace, name)
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
	c := m.(*client.V1Client)
	scope, clusterUid, vmNamespace, vmName, err := utils.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	hapiVM, err := c.GetVirtualMachine(scope, clusterUid, vmNamespace, vmName)
	if err != nil {
		return diag.FromErr(err)
	}
	if hapiVM == nil {
		return diag.FromErr(fmt.Errorf("cannot read virtual machine %s, %s, %s", clusterUid, vmNamespace, vmName))
	}

	// prepare new vm data
	vm, err := virtualmachine.FromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}
	hapiVM, err = convert.ToHapiVm(vm)
	if err != nil {
		return diag.FromErr(err)
	}

	// needed to get context for the cluster
	cluster, err := c.GetCluster(scope, clusterUid)
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

	if _, ok := d.GetOk("vm_action"); ok && d.HasChange("vm_action") {
		stateToChange := d.Get("vm_action").(string)
		resourceVirtualMachineActions(c, ctx, d, stateToChange, scope, clusterUid, vmName, vmNamespace)
	}

	return resourceKubevirtVirtualMachineRead(ctx, d, m)
}

func resourceVirtualMachineActions(c *client.V1Client, ctx context.Context, d *schema.ResourceData, stateToChange string, scope string, clusterUid string, vmName string, vmNamespace string) diag.Diagnostics {
	var diags diag.Diagnostics
	ClusterContext := d.Get("cluster_context").(string)
	// need to add validation status and allowed actions
	// Stopped  - start
	// Paused - restart, resume
	// Running - stop ,restart,pause, migrate
	switch strings.ToLower(stateToChange) {
	//"start", "stop", "restart", "pause", "resume", "migrate"
	case "start":
		err := c.StartVirtualMachine(scope, clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, ClusterContext, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
	case "stop":
		err := c.StopVirtualMachine(scope, clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, ClusterContext, clusterUid, vmName, vmNamespace, diags, c, "update", "Stopped")
		if diags.HasError() {
			return diags
		}
	case "restart":
		err := c.RestartVirtualMachine(scope, clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, ClusterContext, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
	case "pause":
		err := c.PauseVirtualMachine(scope, clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, ClusterContext, clusterUid, vmName, vmNamespace, diags, c, "update", "Paused")
		if diags.HasError() {
			return diags
		}
	case "resume":
		err := c.ResumeVirtualMachine(scope, clusterUid, vmName, vmNamespace)
		if err != nil {
			return diag.FromErr(err)
		}
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, ClusterContext, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
	case "migrate":
		_ = c.MigrateVirtualMachineNodeToNode(scope, clusterUid, vmName, vmNamespace)
		diags, _ = waitForVirtualMachineToTargetState(ctx, d, ClusterContext, clusterUid, vmName, vmNamespace, diags, c, "update", "Running")
		if diags.HasError() {
			return diags
		}
	}
	hapiVM, err := c.GetVirtualMachine(ClusterContext, clusterUid, vmNamespace, vmName)
	if err != nil {
		return diag.FromErr(err)
	}
	if hapiVM == nil {
		return diag.FromErr(fmt.Errorf("cannot read virtual machine after update %s, %s, %s", clusterUid, vmNamespace, vmName))
	}
	return diags
}

func resourceKubevirtVirtualMachineDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	scope, clusterUid, namespace, name, err := utils.IdParts(resourceData.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cli := (meta).(*client.V1Client)

	log.Printf("[INFO] Deleting virtual machine: %#v", name)
	if err := cli.DeleteVirtualMachine(scope, clusterUid, namespace, name); err != nil {
		return diag.FromErr(err)
	}
	diags, _ = waitForVirtualMachineToTargetState(ctx, resourceData, scope, clusterUid, name, namespace, diags, cli, "delete", "Deleted")
	if diags.HasError() {
		return diags
	}
	log.Printf("[INFO] virtual machine %s deleted", name)

	resourceData.SetId("")
	return nil
}
