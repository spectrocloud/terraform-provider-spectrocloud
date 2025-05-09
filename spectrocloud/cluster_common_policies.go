package spectrocloud

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"strings"
)

func toPolicies(d *schema.ResourceData) *models.V1SpectroClusterPolicies {
	return &models.V1SpectroClusterPolicies{
		BackupPolicy: toBackupPolicy(d),
		ScanPolicy:   toScanPolicy(d),
	}
}

func toBackupPolicy(d *schema.ResourceData) *models.V1ClusterBackupConfig {
	if policies, found := d.GetOk("backup_policy"); found {
		policy := policies.([]interface{})[0].(map[string]interface{})

		namespaces := make([]string, 0)
		if policy["namespaces"] != nil {
			if nss, ok := policy["namespaces"]; ok {
				for _, ns := range nss.(*schema.Set).List() {
					namespaces = append(namespaces, ns.(string))
				}
			}
		}

		// Extract and process the policy settings
		includeClusterResourceMode := models.V1IncludeClusterResourceMode("Auto") // Default value
		if policy["include_cluster_resources_mode"] != "" {
			if v, ok := policy["include_cluster_resources_mode"].(string); ok {
				includeClusterResourceMode = convertIncludeResourceMode(v)
			}
		} else if policy["include_cluster_resources"] != nil {
			if include, ok := policy["include_cluster_resources"].(bool); ok {
				if include {
					includeClusterResourceMode = convertIncludeResourceMode("Always")
				} else {
					includeClusterResourceMode = convertIncludeResourceMode("Never")
				}
			}
		}

		return &models.V1ClusterBackupConfig{
			BackupLocationUID:          policy["backup_location_id"].(string),
			BackupPrefix:               policy["prefix"].(string),
			DurationInHours:            int64(policy["expiry_in_hour"].(int)),
			IncludeAllDisks:            policy["include_disks"].(bool),
			IncludeClusterResourceMode: includeClusterResourceMode,
			Namespaces:                 namespaces,
			Schedule: &models.V1ClusterFeatureSchedule{
				ScheduledRunTime: policy["schedule"].(string),
			},
		}
	}
	return nil
}

func convertIncludeResourceMode(m string) (mode models.V1IncludeClusterResourceMode) {
	switch strings.ToLower(m) {
	case "always":
		return models.V1IncludeClusterResourceMode("Always")
	case "never":
		return models.V1IncludeClusterResourceMode("Never")
	case "auto":
		return models.V1IncludeClusterResourceMode("Auto")
	}
	return ""
}

func flattenBackupPolicy(policy *models.V1ClusterBackupConfig, d *schema.ResourceData) []interface{} {
	result := make([]interface{}, 0, 1)
	data := make(map[string]interface{})
	data["schedule"] = policy.Schedule.ScheduledRunTime
	data["backup_location_id"] = policy.BackupLocationUID
	data["prefix"] = policy.BackupPrefix
	data["namespaces"] = policy.Namespaces
	data["expiry_in_hour"] = policy.DurationInHours
	data["include_disks"] = policy.IncludeAllDisks

	if policies, found := d.GetOk("backup_policy"); found {
		bPolicy := policies.([]interface{})[0].(map[string]interface{})
		if bPolicy["include_cluster_resources_mode"] != "" {
			data["include_cluster_resources_mode"] = strings.ToLower(string(policy.IncludeClusterResourceMode))
			data["include_cluster_resources"] = true
		} else {
			data["include_cluster_resources"] = flattenIncludeResourceMode(policy.IncludeClusterResourceMode)
		}
	}
	result = append(result, data)
	return result
}

func flattenIncludeResourceMode(m models.V1IncludeClusterResourceMode) bool {
	return m == models.V1IncludeClusterResourceMode("Always")
}

func updateBackupPolicy(c *client.V1Client, d *schema.ResourceData) error {
	if policy := toBackupPolicy(d); policy != nil {
		//clusterContext := d.Get("context").(string)
		return c.ApplyClusterBackupConfig(d.Id(), policy)
	} else {
		return errors.New("backup policy validation: The backup policy cannot be destroyed. To disable it, set the schedule to an empty string")
	}
}

func toScanPolicy(d *schema.ResourceData) *models.V1ClusterComplianceScheduleConfig {
	if profiles, found := d.GetOk("scan_policy"); found {
		config := &models.V1ClusterComplianceScheduleConfig{}
		policy := profiles.([]interface{})[0].(map[string]interface{})
		if policy["configuration_scan_schedule"] != nil {
			config.KubeBench = &models.V1ClusterComplianceScanKubeBenchScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: policy["configuration_scan_schedule"].(string),
				},
			}
		}
		if policy["penetration_scan_schedule"] != nil {
			config.KubeHunter = &models.V1ClusterComplianceScanKubeHunterScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: policy["penetration_scan_schedule"].(string),
				},
			}
		}
		if policy["conformance_scan_schedule"] != nil {
			config.Sonobuoy = &models.V1ClusterComplianceScanSonobuoyScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: policy["conformance_scan_schedule"].(string),
				},
			}
		}
		return config
	}
	return nil
}

func flattenScanPolicy(driverSpec map[string]models.V1ComplianceScanDriverSpec) []interface{} {
	result := make([]interface{}, 0, 1)
	data := map[string]interface{}{
		"configuration_scan_schedule": "",
		"penetration_scan_schedule":   "",
		"conformance_scan_schedule":   "",
	}

	if v, found := driverSpec["kube-bench"]; found {
		if v.Config.Schedule.ScheduledRunTime == "" {
			data["configuration_scan_schedule"] = ""
		} else {
			data["configuration_scan_schedule"] = v.Config.Schedule.ScheduledRunTime
		}
	}
	if v, found := driverSpec["kube-hunter"]; found {
		if v.Config.Schedule.ScheduledRunTime == "" {
			data["penetration_scan_schedule"] = ""
		} else {
			data["penetration_scan_schedule"] = v.Config.Schedule.ScheduledRunTime
		}
	}
	if v, found := driverSpec["sonobuoy"]; found {
		if v.Config.Schedule.ScheduledRunTime == "" {
			data["conformance_scan_schedule"] = ""
		} else {
			data["conformance_scan_schedule"] = v.Config.Schedule.ScheduledRunTime
		}
	}
	if data["configuration_scan_schedule"] == "" && data["penetration_scan_schedule"] == "" && data["conformance_scan_schedule"] == "" {
		return result
	} else {
		result = append(result, data)
	}
	return result
}

func updateScanPolicy(c *client.V1Client, d *schema.ResourceData) error {
	if policy := toScanPolicy(d); policy != nil || d.HasChange("scan_policy") {
		//ClusterContext := d.Get("context").(string)
		if policy == nil {
			policy = getEmptyScanPolicy()
		}
		return c.ApplyClusterScanConfig(d.Id(), policy)
	}
	return nil
}

func getEmptyScanPolicy() *models.V1ClusterComplianceScheduleConfig {
	scanPolicy := &models.V1ClusterComplianceScheduleConfig{
		KubeBench:  &models.V1ClusterComplianceScanKubeBenchScheduleConfig{Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: ""}},
		KubeHunter: &models.V1ClusterComplianceScanKubeHunterScheduleConfig{Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: ""}},
		Sonobuoy:   &models.V1ClusterComplianceScanSonobuoyScheduleConfig{Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: ""}},
	}
	return scanPolicy
}
