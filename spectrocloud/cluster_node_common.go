package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

var NodeMaintenanceLifecycleStates = []string{
	"Completed",
	"InProgress",
	"Initiated",
	"Failed",
}

type GetMaintenanceStatus func(string, string, string) (*models.V1MachineMaintenanceStatus, error)

type GetNodeStatusMap func(string, string, string) (map[string]models.V1CloudMachineStatus, error)

func waitForNodeMaintenanceCompleted(c *client.V1Client, ctx context.Context, fn GetMaintenanceStatus, ClusterContext, ConfigUID, MachineName, NodeId string) (error, bool) {

	stateConf := &retry.StateChangeConf{
		Delay:      30 * time.Second,
		Pending:    NodeMaintenanceLifecycleStates,
		Target:     []string{"Completed"},
		Refresh:    resourceClusterNodeMaintenanceRefreshFunc(c, fn, ConfigUID, MachineName, NodeId),
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

func resourceClusterNodeMaintenanceRefreshFunc(c *client.V1Client, fn GetMaintenanceStatus, ConfigUID, MachineName, NodeId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		nmStatus, err := c.GetNodeMaintenanceStatus(client.GetMaintenanceStatus(fn), ConfigUID, MachineName, NodeId)
		if err != nil {
			return nil, "", err
		}

		state := nmStatus.State
		log.Printf("Node maintenance state (%s): %s", NodeId, state)

		return nmStatus, state, nil
	}
}

func resourceNodeAction(c *client.V1Client, ctx context.Context, newMachinePool interface{}, fn GetMaintenanceStatus, CloudType, ClusterContext, ConfigUID, MachineName string) error {
	newNodes := newMachinePool.(map[string]interface{})["node"]
	if newNodes != nil {
		for _, n := range newNodes.([]interface{}) {
			node := n.(map[string]interface{})
			nodeMaintenanceStatus, err := c.GetNodeMaintenanceStatus(client.GetMaintenanceStatus(fn), ConfigUID, MachineName, node["node_id"].(string))
			if err != nil {
				return err
			}
			if node["action"] != nodeMaintenanceStatus.Action {
				nm := &models.V1MachineMaintenance{
					Action: node["action"].(string),
				}
				err := c.ToggleMaintenanceOnNode(nm, CloudType, ConfigUID, MachineName, node["node_id"].(string))
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

func flattenNodeMaintenanceStatus(c *client.V1Client, d *schema.ResourceData, fn GetNodeStatusMap, mPools []interface{}, cloudConfigId, ClusterContext string) ([]interface{}, error) {
	_, n := d.GetChange("machine_pool")
	nsMap := make(map[string]interface{})
	machinePoolsList, i, err := getMachinePoolList(n)
	if err != nil {
		return i, err
	}

	for _, mp := range machinePoolsList {
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

func getMachinePoolList(n interface{}) ([]interface{}, []interface{}, error) {
	var machinePoolsList []interface{}

	// Check if n is of type *schema.Set
	if set, ok := n.(*schema.Set); ok {
		machinePoolsList = set.List()
	} else if list, ok := n.([]interface{}); ok {
		// If n is already a slice of interfaces
		machinePoolsList = list
	} else {
		// Handle error: n is neither *schema.Set nor []interface{}
		return nil, nil, fmt.Errorf("unexpected type for n: %T", n)
	}
	return machinePoolsList, nil, nil
}
