package spectrocloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/robfig/cron"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func updateClusterOsPatchConfig(c *client.V1Client, d *schema.ResourceData) error {
	machineConfig := toMachineManagementConfig(d)
	clusterContext := d.Get("context").(string)
	err := ValidateContext(clusterContext)
	if err != nil {
		return err
	}
	if machineConfig.OsPatchConfig != nil {
		return c.UpdateClusterOsPatchConfig(d.Id(), clusterContext, toUpdateOsPatchEntityClusterRbac(machineConfig.OsPatchConfig))
	} else {
		return c.UpdateClusterOsPatchConfig(d.Id(), clusterContext, toUpdateOsPatchEntityClusterRbac(getDefaultOsPatchConfig().OsPatchConfig))
	}
}

func getDefaultOsPatchConfig() *models.V1MachineManagementConfig {
	return &models.V1MachineManagementConfig{
		OsPatchConfig: &models.V1OsPatchConfig{
			PatchOnBoot:      false,
			RebootIfRequired: false,
		},
	}
}

func toUpdateOsPatchEntityClusterRbac(config *models.V1OsPatchConfig) *models.V1OsPatchEntity {
	return &models.V1OsPatchEntity{
		OsPatchConfig: config,
	}
}

func toOsPatchConfig(d *schema.ResourceData) *models.V1OsPatchConfig {
	osPatchOnBoot := false
	_, isOsPatchOnBoot := d.GetOk("os_patch_on_boot")
	_, isOsPatchOnSchedule := d.GetOk("os_patch_schedule")
	_, isOsPatchAfter := d.GetOk("os_patch_after")
	if isOsPatchOnBoot || isOsPatchOnSchedule || isOsPatchAfter {
		if d.Get("os_patch_on_boot") != nil {
			osPatchOnBoot = d.Get("os_patch_on_boot").(bool)
		}
		osPatchOnSchedule := d.Get("os_patch_schedule").(string)
		osPatchAfter := d.Get("os_patch_after").(string)
		if osPatchOnBoot || len(osPatchOnSchedule) > 0 || len(osPatchAfter) > 0 {
			osPatchConfig := &models.V1OsPatchConfig{}
			if osPatchOnBoot {
				osPatchConfig.PatchOnBoot = osPatchOnBoot
			}
			if len(osPatchOnSchedule) > 0 {
				osPatchConfig.Schedule = osPatchOnSchedule
			}
			if len(osPatchAfter) > 0 {
				patchAfter, _ := time.Parse(time.RFC3339, osPatchAfter)
				osPatchConfig.OnDemandPatchAfter = models.V1Time(patchAfter)
			} else {
				//setting Zero time in request
				zeroTime, _ := time.Parse(time.RFC3339, "0001-01-01T00:00:00.000Z")
				osPatchConfig.OnDemandPatchAfter = models.V1Time(zeroTime)
			}
			return osPatchConfig
		}
	}
	return nil
}

func validateOsPatchSchedule(data interface{}, _ cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if data != nil {
		if _, err := cron.ParseStandard(data.(string)); err != nil {
			wrappedErr := fmt.Errorf("os patch schedule is invalid. Please see https://en.wikipedia.org/wiki/Cron for valid cron syntax: %w", err)
			return diag.FromErr(wrappedErr)
		}
	}
	return diags
}

func validateOsPatchOnDemandAfter(data interface{}, _ cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if data != nil {
		if patchTime, err := time.Parse(time.RFC3339, data.(string)); err != nil {
			wrappedErr := fmt.Errorf("time for 'os_patch_after' is invalid. Please follow RFC3339 Date and Time Standards. Eg 2021-01-01T00:00:00.000Z : %w", err)
			return diag.FromErr(wrappedErr)
		} else {
			if time.Now().After(patchTime.Add(10 * time.Minute)) {
				wrappedErr := fmt.Errorf("valid timestamp is timestamp which is 10 mins ahead of current timestamp. Eg any timestamp ahead of %v", time.Now().Add(10*time.Minute).Format(time.RFC3339))
				return diag.FromErr(wrappedErr)
			}
		}
	}

	return diags
}
