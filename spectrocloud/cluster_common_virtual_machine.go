package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// fix as needed with real statuses
var resourceVirtualMachineCreatePendingStates = []string{
	"Creating",
	"Created",
	"Running",
}

func waitForVirtualMachine(ctx context.Context, d *schema.ResourceData, cluster_uid string, vm_uid string, diags diag.Diagnostics, c *client.V1Client, state string) (diag.Diagnostics, bool) {
	cluster, err := c.GetCluster(cluster_uid)
	if err != nil {
		return diags, true
	}

	if _, found := cluster.Metadata.Labels["skip_vms"]; found {
		return diags, true
	}

	stateConf := &resource.StateChangeConf{
		Pending:    resourceVirtualMachineCreatePendingStates,
		Target:     []string{"True"},
		Refresh:    resourceVirtualMachineStateRefreshFunc(c, cluster_uid, vm_uid),
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

func waitForVirtualMachineCreation(ctx context.Context, d *schema.ResourceData, cluster_uid string, profile_uid string, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	return waitForVirtualMachine(ctx, d, cluster_uid, profile_uid, diags, c, schema.TimeoutCreate)
}

func waitForVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, cluster_uid string, profile_uid string, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	return waitForVirtualMachine(ctx, d, cluster_uid, profile_uid, diags, c, schema.TimeoutUpdate)
}

func resourceVirtualMachineStateRefreshFunc(c *client.V1Client, cluster_uid string, vm_uid string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := c.GetCluster(cluster_uid)
		if err != nil {
			return nil, "", err
		} else if cluster == nil {
			return nil, "Deleted", nil
		}

		//TODO: wait for nodes to be ready

		return cluster, "True", nil
	}
}

// TODO: implement it.
func resourceVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	return diags
}
