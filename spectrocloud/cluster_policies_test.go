package spectrocloud

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestFlattenScanPolicy(t *testing.T) {
	driverSpec := map[string]models.V1ComplianceScanDriverSpec{
		"kube-bench": {
			Config: &models.V1ComplianceScanConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: "daily",
				},
			},
		},
		"kube-hunter": {
			Config: &models.V1ComplianceScanConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: "hourly",
				},
			},
		},
		"sonobuoy": {
			Config: &models.V1ComplianceScanConfig{
				Schedule: &models.V1ClusterFeatureSchedule{
					ScheduledRunTime: "weekly",
				},
			},
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"configuration_scan_schedule": "daily",
			"penetration_scan_schedule":   "hourly",
			"conformance_scan_schedule":   "weekly",
		},
	}

	result := flattenScanPolicy(driverSpec)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Result does not match expected. Got %v, expected %v", result, expected)
	}
}

func TestGetEmptyScanPolicy(t *testing.T) {
	result := getEmptyScanPolicy()

	expected := &models.V1ClusterComplianceScheduleConfig{
		KubeBench:  &models.V1ClusterComplianceScanKubeBenchScheduleConfig{Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: ""}},
		KubeHunter: &models.V1ClusterComplianceScanKubeHunterScheduleConfig{Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: ""}},
		Sonobuoy:   &models.V1ClusterComplianceScanSonobuoyScheduleConfig{Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: ""}},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Result does not match expected. Got %v, expected %v", result, expected)
	}
}

func TestFlattenBackupPolicy(t *testing.T) {
	policy := &models.V1ClusterBackupConfig{
		Schedule:                &models.V1ClusterFeatureSchedule{ScheduledRunTime: "daily"},
		BackupLocationUID:       "location-123",
		BackupPrefix:            "backup-prefix",
		Namespaces:              []string{"namespace1", "namespace2"},
		DurationInHours:         24,
		IncludeAllDisks:         true,
		IncludeClusterResources: true,
	}

	expected := []interface{}{
		map[string]interface{}{
			"schedule":                  "daily",
			"backup_location_id":        "location-123",
			"prefix":                    "backup-prefix",
			"namespaces":                []string{"namespace1", "namespace2"},
			"expiry_in_hour":            int64(24),
			"include_disks":             true,
			"include_cluster_resources": true,
		},
	}

	result := flattenBackupPolicy(policy)
	assert.Equal(t, expected, result)
}

func TestToBackupPolicy(t *testing.T) {
	// Create a ResourceData to simulate Terraform state
	resourceData := resourceClusterAws().TestResourceData()
	backupPolicy := []interface{}{
		map[string]interface{}{
			"backup_location_id":        "location-123",
			"prefix":                    "backup-prefix",
			"expiry_in_hour":            24,
			"include_disks":             true,
			"include_cluster_resources": true,
			"namespaces":                []interface{}{"namespace1"},
			"schedule":                  "daily",
		},
	}
	err := resourceData.Set("backup_policy", backupPolicy)
	if err != nil {
		return
	}

	result := toBackupPolicy(resourceData)

	expected := &models.V1ClusterBackupConfig{
		BackupLocationUID:       "location-123",
		BackupPrefix:            "backup-prefix",
		DurationInHours:         24,
		IncludeAllDisks:         true,
		IncludeClusterResources: true,
		Namespaces:              []string{"namespace1"},
		Schedule: &models.V1ClusterFeatureSchedule{
			ScheduledRunTime: "daily",
		},
	}

	assert.Equal(t, expected, result)
}

func TestToScanPolicy(t *testing.T) {
	// Create a ResourceData to simulate Terraform state
	resourceData := resourceClusterAws().TestResourceData()

	scanPolicy := []interface{}{
		map[string]interface{}{
			"configuration_scan_schedule": "daily",
			"penetration_scan_schedule":   "hourly",
			"conformance_scan_schedule":   "weekly",
		},
	}
	err := resourceData.Set("scan_policy", scanPolicy)
	if err != nil {
		return
	}
	result := toScanPolicy(resourceData)

	expected := &models.V1ClusterComplianceScheduleConfig{
		KubeBench: &models.V1ClusterComplianceScanKubeBenchScheduleConfig{
			Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: "daily"},
		},
		KubeHunter: &models.V1ClusterComplianceScanKubeHunterScheduleConfig{
			Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: "hourly"},
		},
		Sonobuoy: &models.V1ClusterComplianceScanSonobuoyScheduleConfig{
			Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: "weekly"},
		},
	}

	assert.Equal(t, expected, result)
}

func TestToPolicies(t *testing.T) {
	// Create a ResourceData to simulate Terraform state
	resourceData := resourceClusterAws().TestResourceData()
	backupPolicy := []interface{}{
		map[string]interface{}{
			"backup_location_id":        "location-123",
			"prefix":                    "backup-prefix",
			"expiry_in_hour":            24,
			"include_disks":             true,
			"include_cluster_resources": true,
			"namespaces":                []interface{}{"namespace1"},
			"schedule":                  "daily",
		},
	}
	err := resourceData.Set("backup_policy", backupPolicy)
	if err != nil {
		return
	}
	scanPolicy := []interface{}{
		map[string]interface{}{
			"configuration_scan_schedule": "daily",
			"penetration_scan_schedule":   "hourly",
			"conformance_scan_schedule":   "weekly",
		},
	}
	err = resourceData.Set("scan_policy", scanPolicy)
	if err != nil {
		return
	}

	result := toPolicies(resourceData)

	expected := &models.V1SpectroClusterPolicies{
		BackupPolicy: &models.V1ClusterBackupConfig{
			BackupLocationUID:       "location-123",
			BackupPrefix:            "backup-prefix",
			DurationInHours:         24,
			IncludeAllDisks:         true,
			IncludeClusterResources: true,
			Namespaces:              []string{"namespace1"},
			Schedule: &models.V1ClusterFeatureSchedule{
				ScheduledRunTime: "daily",
			},
		},
		ScanPolicy: &models.V1ClusterComplianceScheduleConfig{
			KubeBench: &models.V1ClusterComplianceScanKubeBenchScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: "daily"},
			},
			KubeHunter: &models.V1ClusterComplianceScanKubeHunterScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: "hourly"},
			},
			Sonobuoy: &models.V1ClusterComplianceScanSonobuoyScheduleConfig{
				Schedule: &models.V1ClusterFeatureSchedule{ScheduledRunTime: "weekly"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestValidateContext(t *testing.T) {
	// Test valid context
	err := ValidateContext("project")
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	err = ValidateContext("tenant")
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Test invalid context
	err = ValidateContext("invalid")
	expectedError := fmt.Errorf("invalid Context set - invalid")
	if err == nil || err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
	}
}
