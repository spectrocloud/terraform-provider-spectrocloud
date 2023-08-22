package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

func waitForNodeMaintenanceCompleted(c *client.V1Client, ctx context.Context, CloudType string, ClusterContext string, ConfigUID string, MachineName string, NodeId string) (error, bool) {

	stateConf := &retry.StateChangeConf{
		Delay:      30 * time.Second,
		Pending:    NodeMaintenanceLifecycleStates,
		Target:     []string{"Completed"},
		Refresh:    resourceClusterNodeMaintenanceRefreshFunc(c, CloudType, ClusterContext, ConfigUID, MachineName, NodeId),
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

func resourceClusterNodeMaintenanceRefreshFunc(c *client.V1Client, CloudType string, ClusterContext string, ConfigUID string, MachineName string, NodeId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		nmStatus, err := c.GetNodeMaintenanceStatus(CloudType, ClusterContext, ConfigUID, MachineName, NodeId)
		if err != nil {
			return nil, "", err
		}

		state := nmStatus.State
		log.Printf("Node maintenance state (%s): %s", NodeId, state)

		return nmStatus, state, nil
	}
}

func resourceNodeAction(c *client.V1Client, ctx context.Context, newMachinePool interface{}, cloudType string, ClusterContext string, ConfigUID string, MachineName string) error {

	newNodes := newMachinePool.(map[string]interface{})["node"]
	if newNodes != nil {
		for _, n := range newNodes.([]interface{}) {
			node := n.(map[string]interface{})
			nodeMaintenanceStatus, err := c.GetNodeMaintenanceStatus(cloudType, ClusterContext, ConfigUID, MachineName, node["node_id"].(string))
			if err != nil {
				return err
			}
			if node["action"] != nodeMaintenanceStatus.Action {
				nm := &models.V1MachineMaintenance{
					Action: node["action"].(string),
				}
				err := c.ToggleMaintenanceOnNode(nm, cloudType, ClusterContext, ConfigUID, MachineName, node["node_id"].(string))
				if err != nil {
					return err
				}
				err, isError := waitForNodeMaintenanceCompleted(c, ctx, cloudType, ClusterContext, ConfigUID, MachineName, node["node_id"].(string))
				if isError {
					return err
				}
			}
		}
	}
	return nil
}
