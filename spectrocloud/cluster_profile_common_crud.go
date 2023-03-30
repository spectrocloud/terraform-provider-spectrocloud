package spectrocloud

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/spectrocloud/palette-sdk-go/client"
)

var resourceClusterProfileUpdatePendingStates = []string{
	"false",
}

func waitForProfileDownload(ctx context.Context, c *client.V1Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    resourceClusterProfileUpdatePendingStates,
		Target:     []string{"true"}, // canBeApplied=true
		Refresh:    resourceClusterProfileStateRefreshFunc(c, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func resourceClusterProfileStateRefreshFunc(c *client.V1Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := c.GetCluster(id)
		if err != nil {
			return nil, "", err
		} else if cluster == nil {
			return nil, "Cluster deleted", nil
		}

		state := strconv.FormatBool(cluster.Status.SpcApply.CanBeApplied)
		log.Printf("Cluster SpcApply state (%s): %s", id, state)

		return cluster, state, nil
	}
}
