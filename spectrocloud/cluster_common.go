package spectrocloud

import (
	"fmt"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/robfig/cron"
	"github.com/spectrocloud/hapi/models"
)

var (
	DefaultDiskType = "Standard_LRS"
	DefaultDiskSize = 60
)

func toTags(d *schema.ResourceData) map[string]string {
	tags := make(map[string]string)
	if d.Get("tags") != nil {
		for _, t := range d.Get("tags").(*schema.Set).List() {
			tag := t.(string)
			if strings.Contains(tag, ":") {
				tags[strings.Split(tag, ":")[0]] = strings.Split(tag, ":")[1]
			} else {
				tags[tag] = "spectro__tag"
			}
		}
		return tags
	} else {
		return nil
	}
}

func toNtpServers(in map[string]interface{}) []string {
	servers := make([]string, 0, 1)
	if _, ok := in["ntp_servers"]; ok {
		for _, t := range in["ntp_servers"].(*schema.Set).List() {
			ntp := t.(string)
			servers = append(servers, ntp)
		}
	}
	return servers
}

func flattenTags(labels map[string]string) []interface{} {
	tags := make([]interface{}, 0)
	if len(labels) > 0 {
		for k, v := range labels {
			if v == "spectro__tag" {
				tags = append(tags, k)
			} else {
				tags = append(tags, fmt.Sprintf("%s:%s", k, v))
			}
		}
		return tags
	} else {
		return nil
	}
}

func toClusterConfig(d *schema.ResourceData) *models.V1ClusterConfigEntity {
	return &models.V1ClusterConfigEntity{
		MachineManagementConfig: toMachineManagementConfig(d),
		Resources:               toClusterResourceConfig(d),
	}
}

func toMachineManagementConfig(d *schema.ResourceData) *models.V1MachineManagementConfig {
	return &models.V1MachineManagementConfig{
		OsPatchConfig: toOsPatchConfig(d),
	}
}

func toClusterResourceConfig(d *schema.ResourceData) *models.V1ClusterResourcesEntity {
	return &models.V1ClusterResourcesEntity{
		Namespaces: toClusterNamespaces(d),
		Rbacs:      toClusterRBACs(d),
	}
}

func toOsPatchConfig(d *schema.ResourceData) *models.V1OsPatchConfig {
	osPatchOnBoot := d.Get("os_patch_on_boot").(bool)
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
	return nil
}

func validateOsPatchSchedule(data interface{}, _ cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if data != nil {
		if _, err := cron.ParseStandard(data.(string)); err != nil {
			return diag.FromErr(errors.Wrap(err, "os patch schedule is invalid. Please see https://en.wikipedia.org/wiki/Cron for valid cron syntax"))
		}
	}
	return diags
}

func validateOsPatchOnDemandAfter(data interface{}, _ cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if data != nil {
		if patchTime, err := time.Parse(time.RFC3339, data.(string)); err != nil {
			return diag.FromErr(errors.Wrap(err, "time for 'os_patch_after' is invalid. Please follow RFC3339 Date and Time Standards. Eg 2021-01-01T00:00:00.000Z "))
		} else {
			if time.Now().After(patchTime.Add(10 * time.Minute)) {
				return diag.FromErr(fmt.Errorf("valid timestamp is timestamp which is 10 mins ahead of current timestamp. Eg any timestamp ahead of %v", time.Now().Add(10*time.Minute).Format(time.RFC3339)))
			}
		}
	}

	return diags
}

//read common fields like kubeconfig, tags, backup policy, scan policy, cluster_rbac_binding, namespaces
func readCommonFields(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) (diag.Diagnostics, bool) {
	kubecfg, err := c.GetClusterKubeConfig(d.Id())
	if err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("kubeconfig", kubecfg); err != nil {
		return diag.FromErr(err), true
	}

	if err := d.Set("tags", flattenTags(cluster.Metadata.Labels)); err != nil {
		return diag.FromErr(err), true
	}

	if policy, err := c.GetClusterBackupConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if policy != nil && policy.Spec.Config != nil {
		if err := d.Set("backup_policy", flattenBackupPolicy(policy.Spec.Config)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if policy, err := c.GetClusterScanConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if policy != nil && policy.Spec.DriverSpec != nil {
		if err := d.Set("scan_policy", flattenScanPolicy(policy.Spec.DriverSpec)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if rbac, err := c.GetClusterRbacConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if rbac != nil && rbac.Items != nil {
		if err := d.Set("cluster_rbac_binding", flattenClusterRBAC(rbac.Items)); err != nil {
			return diag.FromErr(err), true
		}
	}

	if namespace, err := c.GetClusterNamespaceConfig(d.Id()); err != nil {
		return diag.FromErr(err), true
	} else if namespace != nil && namespace.Items != nil {
		if err := d.Set("namespaces", flattenClusterNamespaces(namespace.Items)); err != nil {
			return diag.FromErr(err), true
		}
	}
	return diag.Diagnostics{}, false
}

// update common fields like namespaces, cluster_rbac_binding, cluster_profile, backup_policy, scan_policy
func updateCommonFields(d *schema.ResourceData, c *client.V1Client) (diag.Diagnostics, bool) {
	if d.HasChange("namespaces") {
		if err := updateClusterNamespaces(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("cluster_rbac_binding") {
		if err := updateClusterRBAC(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("os_patch_on_boot") || d.HasChange("os_patch_schedule") || d.HasChange("os_patch_after") {
		if err := updateClusterOsPatchConfig(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChanges("cluster_profile", "packs", "manifests") {
		if err := updateProfiles(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("backup_policy") {
		if err := updateBackupPolicy(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}

	if d.HasChange("scan_policy") {
		if err := updateScanPolicy(c, d); err != nil {
			return diag.FromErr(err), true
		}
	}
	return diag.Diagnostics{}, false
}
