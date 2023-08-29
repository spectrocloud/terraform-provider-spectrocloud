package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"log"
	"time"
)

var NodeMaintenanceLifecycleStates = []string{
	"Completed",
	"InProgress",
	"Initiated",
	"Failed",
}

type GetMaintenanceStatus func(string, string, string, string) (*models.V1MachineMaintenanceStatus, error)

type GetNodeStatusMap func(string, string, string) (map[string]models.V1CloudMachineStatus, error)

func waitForNodeMaintenanceCompleted(c *client.V1Client, ctx context.Context, fn GetMaintenanceStatus, ClusterContext string, ConfigUID string, MachineName string, NodeId string) (error, bool) {

	stateConf := &retry.StateChangeConf{
		Delay:      30 * time.Second,
		Pending:    NodeMaintenanceLifecycleStates,
		Target:     []string{"Completed"},
		Refresh:    resourceClusterNodeMaintenanceRefreshFunc(c, fn, ClusterContext, ConfigUID, MachineName, NodeId),
		Timeout:    30 * time.Minute,
		MinTimeout: 10 * time.Second,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return err, true
	}
	return nil, false
}

func resourceClusterNodeMaintenanceRefreshFunc(c *client.V1Client, fn GetMaintenanceStatus, ClusterContext string, ConfigUID string, MachineName string, NodeId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		nmStatus, err := c.GetNodeMaintenanceStatus(client.GetMaintenanceStatus(fn), ClusterContext, ConfigUID, MachineName, NodeId)
		if err != nil {
			return nil, "", err
		}

		state := nmStatus.State
		log.Printf("Node maintenance state (%s): %s", NodeId, state)

		return nmStatus, state, nil
	}
}

func resourceNodeAction(c *client.V1Client, ctx context.Context, newMachinePool interface{}, fn GetMaintenanceStatus, CloudType string, ClusterContext string, ConfigUID string, MachineName string) error {
	newNodes := newMachinePool.(map[string]interface{})["node"]
	if newNodes != nil {
		for _, n := range newNodes.([]interface{}) {
			node := n.(map[string]interface{})
			nodeMaintenanceStatus, err := c.GetNodeMaintenanceStatus(client.GetMaintenanceStatus(fn), ClusterContext, ConfigUID, MachineName, node["node_id"].(string))
			if err != nil {
				return err
			}
			if node["action"] != nodeMaintenanceStatus.Action {
				nm := &models.V1MachineMaintenance{
					Action: node["action"].(string),
				}
				err := c.ToggleMaintenanceOnNode(nm, CloudType, ClusterContext, ConfigUID, MachineName, node["node_id"].(string))
				if err != nil {
					return err
				}
				err, isError := waitForNodeMaintenanceCompleted(c, ctx, fn, ClusterContext, ConfigUID, MachineName, node["node_id"].(string))
				if isError {
					return err
				}
			}
		}
	}
	return nil
}

func flattenNodeMaintenanceStatus(c *client.V1Client, d *schema.ResourceData, fn GetNodeStatusMap, mPools []interface{}, cloudConfigId string, ClusterContext string) ([]interface{}, error) {
	_, n := d.GetChange("machine_pool")
	nsMap := make(map[string]interface{})
	for _, mp := range n.(*schema.Set).List() {
		machinePool := mp.(map[string]interface{})
		nsMap[machinePool["name"].(string)] = machinePool
	}

	for i, mp := range mPools {
		m := mp.(map[string]interface{})
		// For handling unit test
		if _, ok := nsMap[m["name"].(string)]; !ok {
			return mPools, nil
		}

		newNodeList := nsMap[m["name"].(string)].(map[string]interface{})["node"].([]interface{})
		if len(newNodeList) > 0 {
			var nodes []interface{}
			nodesStatus, err := fn(cloudConfigId, m["name"].(string), ClusterContext)
			if err != nil {
				return nil, err
			}
			for key, value := range nodesStatus {
				for _, newNode := range newNodeList {
					if newNode.(map[string]interface{})["node_id"] == key {
						nodes = append(nodes, c.GetNodeValue(key, value.MaintenanceStatus.Action))
					}
				}
			}
			if nodes != nil {
				mPools[i].(map[string]interface{})["node"] = nodes
			}
		}
	}
	return mPools, nil
}
