package spectrocloud

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	"github.com/spectrocloud/hapi/models"
)

func TestToAdditionalNodePoolLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name:     "Nil additional_labels",
			input:    map[string]interface{}{"additional_labels": nil},
			expected: map[string]string{},
		},
		{
			name:     "Empty additional_labels",
			input:    map[string]interface{}{"additional_labels": map[string]interface{}{}},
			expected: map[string]string{},
		},
		{
			name: "Valid additional_labels",
			input: map[string]interface{}{
				"additional_labels": map[string]interface{}{
					"label1": "value1",
					"label2": "value2",
				},
			},
			expected: map[string]string{
				"label1": "value1",
				"label2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toAdditionalNodePoolLabels(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToClusterTaints(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected []*models.V1Taint
	}{
		{
			name:     "Nil taints",
			input:    map[string]interface{}{"taints": nil},
			expected: nil,
		},
		{
			name:     "Empty taints",
			input:    map[string]interface{}{"taints": []interface{}{}},
			expected: []*models.V1Taint{},
		},
		{
			name: "Valid taints",
			input: map[string]interface{}{
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "key1",
						"value":  "value1",
						"effect": "NoSchedule",
					},
					map[string]interface{}{
						"key":    "key2",
						"value":  "value2",
						"effect": "PreferNoSchedule",
					},
				},
			},
			expected: []*models.V1Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: "NoSchedule",
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: "PreferNoSchedule",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toClusterTaints(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToClusterTaint(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1Taint
	}{
		{
			name: "Valid cluster taint",
			input: map[string]interface{}{
				"key":    "key1",
				"value":  "value1",
				"effect": "NoSchedule",
			},
			expected: &models.V1Taint{
				Key:    "key1",
				Value:  "value1",
				Effect: "NoSchedule",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toClusterTaint(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFlattenClusterTaints(t *testing.T) {
	taint1 := &models.V1Taint{
		Key:    "key1",
		Value:  "value1",
		Effect: "NoSchedule",
	}
	taint2 := &models.V1Taint{
		Key:    "key2",
		Value:  "value2",
		Effect: "PreferNoSchedule",
	}

	tests := []struct {
		name     string
		input    []*models.V1Taint
		expected []interface{}
	}{
		{
			name:     "Empty items",
			input:    []*models.V1Taint{},
			expected: []interface{}{},
		},
		{
			name:  "Valid taints",
			input: []*models.V1Taint{taint1, taint2},
			expected: []interface{}{
				map[string]interface{}{
					"key":    "key1",
					"value":  "value1",
					"effect": "NoSchedule",
				},
				map[string]interface{}{
					"key":    "key2",
					"value":  "value2",
					"effect": "PreferNoSchedule",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenClusterTaints(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFlattenAdditionalLabelsAndTaints(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		taints   []*models.V1Taint
		expected map[string]interface{}
	}{
		{
			name:     "Empty labels and taints",
			labels:   make(map[string]string),
			taints:   []*models.V1Taint{},
			expected: map[string]interface{}{"additional_labels": map[string]interface{}{}},
		},
		{
			name:   "Non-empty labels",
			labels: map[string]string{"label1": "value1", "label2": "value2"},
			taints: []*models.V1Taint{},
			expected: map[string]interface{}{
				"additional_labels": map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
			},
		},
		{
			name:   "Non-empty labels and taints",
			labels: map[string]string{"label1": "value1", "label2": "value2"},
			taints: []*models.V1Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: "NoSchedule",
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: "PreferNoSchedule",
				},
			},
			expected: map[string]interface{}{
				"additional_labels": map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "key1",
						"value":  "value1",
						"effect": "NoSchedule",
					},
					map[string]interface{}{
						"key":    "key2",
						"value":  "value2",
						"effect": "PreferNoSchedule",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oi := make(map[string]interface{})
			FlattenAdditionalLabelsAndTaints(tt.labels, tt.taints, oi)
			if !reflect.DeepEqual(oi, tt.expected) {
				t.Logf("Expected: %#v\n", tt.expected)
				t.Logf("Actual: %#v\n", oi)
				t.Errorf("Test %s failed. Expected %#v, got %#v", tt.name, tt.expected, oi)
			}
		})
	}
}

func TestUpdateClusterRBAC(t *testing.T) {
	d := resourceClusterVsphere().TestResourceData()

	// Case 1: rbacs context is invalid
	d.Set("context", "invalid")
	err := updateClusterRBAC(nil, d)
	if err == nil || err.Error() != "invalid Context set - invalid" {
		t.Errorf("Expected 'invalid Context set - invalid', got %v", err)
	}
}

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
	resourceData.Set("backup_policy", backupPolicy)

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
	resourceData.Set("scan_policy", scanPolicy)
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
	resourceData.Set("backup_policy", backupPolicy)
	scanPolicy := []interface{}{
		map[string]interface{}{
			"configuration_scan_schedule": "daily",
			"penetration_scan_schedule":   "hourly",
			"conformance_scan_schedule":   "weekly",
		},
	}
	resourceData.Set("scan_policy", scanPolicy)

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
