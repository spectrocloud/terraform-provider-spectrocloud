package spectrocloud

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

var resourceApplicationCreatePendingStates = []string{
	"Tier:Error",
	"PackPending",
	"Tier:NotReady",
	"Application:NotReady",
	"Application:Peding",
}

func waitForApplication(ctx context.Context, d *schema.ResourceData, diags diag.Diagnostics, c *client.V1Client, state string) (diag.Diagnostics, bool) {
	application, err := c.GetApplication(d.Id())
	if err != nil {
		return diags, true
	}

	if _, found := application.Metadata.Labels["skip_apps"]; found {
		return diags, true
	}

	stateConf := &retry.StateChangeConf{
		Pending:    resourceApplicationCreatePendingStates,
		Target:     []string{"True"},
		Refresh:    resourceApplicationStateRefreshFunc(c, d, 5, 60),
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

func waitForApplicationCreation(ctx context.Context, d *schema.ResourceData, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	return waitForApplication(ctx, d, diags, c, schema.TimeoutCreate)
}

func waitForApplicationUpdate(ctx context.Context, d *schema.ResourceData, diags diag.Diagnostics, c *client.V1Client) (diag.Diagnostics, bool) {
	return waitForApplication(ctx, d, diags, c, schema.TimeoutUpdate)
}

func resourceApplicationStateRefreshFunc(c *client.V1Client, d *schema.ResourceData, retryAttempts int, duration int) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		application, err := c.GetApplication(d.Id())
		if err != nil {
			return nil, "", err
		} else if application == nil {
			return nil, "Deleted", nil
		}

		for _, tier_status := range application.Status.AppTiers {
			log.Printf("Cluster (%s): tier:%s, condition status:%s", d.Id(), tier_status.Name, *tier_status.Condition.Status)
			if *tier_status.Condition.Type == "Error" {
				// invoke recursive call h.retryAttempts number of times
				if retryAttempts > 0 {
					time.Sleep(time.Duration(duration) * time.Second)
					return resourceApplicationStateRefreshFunc(c, d, retryAttempts-1, duration)()
				} else {
					return application, "Tier:Error", errors.New(tier_status.Condition.Message)
				}
			}
			if *tier_status.Condition.Status != "True" || *tier_status.Condition.Type != "Ready" {
				return application, "Tier:NotReady", nil
			}
		}

		if application.Status.State != "Deployed" {
			return application, "Application:NotReady", nil
		}

		return application, "True", nil
	}
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	configList := d.Get("config")
	c := getV1ClientWithResourceContext(m, "")
	if configList.([]interface{})[0] != nil {
		config := configList.([]interface{})[0].(map[string]interface{})
		resourceContext := config["cluster_context"].(string)
		c = getV1ClientWithResourceContext(m, resourceContext)
	}
	var diags diag.Diagnostics
	err := c.DeleteApplication(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
