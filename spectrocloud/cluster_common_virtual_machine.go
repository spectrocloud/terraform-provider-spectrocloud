package spectrocloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/spectrocloud/palette-sdk-go/api/apiutil/transport"
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

var resourceVirtualMachineCreatePendingStates = []string{
	"Stopped",
	"Starting",
	"Creating",
	"Provisioning",
	"Created",
	"WaitingForVolumeBinding",
	"Running",
	// Restart|Stop
	"Stopping",
	// Pause
	"Pausing",
	// Migration
	"Migrating",
	//Deleting VM
	"Terminating",
	"Deleted",
}

func waitForVirtualMachineToTargetState(ctx context.Context, d *schema.ResourceData, clusterUid, vmName, namespace string, diags diag.Diagnostics, c *client.V1Client, state, targetState string) (diag.Diagnostics, bool) {
	vm, err := c.GetVirtualMachine(clusterUid, namespace, vmName)
	if err != nil {
		return diags, true
	}
	if vm == nil {
		return diag.FromErr(fmt.Errorf("virtual machine not found when waiting for state %s, %s, %s", clusterUid, namespace, vmName)), true
	}

	if _, found := vm.Metadata.Labels["skip_vms"]; found {
		return diags, true
	}

	stateConf := &retry.StateChangeConf{
		Pending:    resourceVirtualMachineCreatePendingStates,
		Target:     []string{targetState},
		Refresh:    resourceVirtualMachineStateRefreshFunc(c, clusterUid, vmName, namespace),
		Timeout:    d.Timeout(state) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err), true
	}
	return nil, false
}

func resourceVirtualMachineStateRefreshFunc(c *client.V1Client, clusterUid, vmName, vmNamespace string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vm, err := c.GetVirtualMachine(clusterUid, vmNamespace, vmName)
		if err != nil {
			if err.(*transport.TransportError).HttpCode == 500 && strings.Contains(err.(*transport.TransportError).Payload.Message, fmt.Sprintf("Failed to get virtual machine '%s'", vmName)) {
				emptyVM := &models.V1ClusterVirtualMachine{}
				return emptyVM, "Deleted", nil
			} else {
				return nil, "", err
			}
		}
		if vm == nil {
			emptyVM := &models.V1ClusterVirtualMachine{}
			return emptyVM, "", nil
		}
		return vm, vm.Status.PrintableStatus, nil
	}
}
