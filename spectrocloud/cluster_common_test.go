package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/client"
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

func TestRepaveApprovalCheck(t *testing.T) {

	d := resourceClusterAws().TestResourceData()
	d.Set("review_repave_state", "Approved")
	d.Set("context", "tenant")
	d.SetId("TestclusterUID")

	m := &client.V1Client{
		ApproveClusterRepaveFn: func(context, clusterUID string) error {
			return nil
		},
		GetClusterFn: func(context, clusterUID string) (*models.V1SpectroCluster, error) {
			return &models.V1SpectroCluster{
				APIVersion: "",
				Kind:       "",
				Metadata:   nil,
				Spec:       nil,
				Status: &models.V1SpectroClusterStatus{
					Repave: &models.V1ClusterRepaveStatus{
						State: "Approved",
					},
				},
			}, nil
		},
		GetRepaveReasonsFn: func(context, clusterUID string) ([]string, error) {
			var reason []string
			reason = append(reason, "PackValuesUpdated")
			return reason, nil
		},
	}

	// Test case where repave state is pending and approve_system_repave is true
	err := validateSystemRepaveApproval(d, m)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	// Test case where repave state is pending and approve_system_repave is false
	m = &client.V1Client{
		ApproveClusterRepaveFn: func(context, clusterUID string) error {
			return nil
		},
		GetClusterFn: func(context, clusterUID string) (*models.V1SpectroCluster, error) {
			return &models.V1SpectroCluster{
				APIVersion: "",
				Kind:       "",
				Metadata:   nil,
				Spec:       nil,
				Status: &models.V1SpectroClusterStatus{
					Repave: &models.V1ClusterRepaveStatus{
						State: "Pending",
					},
				},
			}, nil
		},
		GetRepaveReasonsFn: func(context, clusterUID string) ([]string, error) {
			var reason []string
			reason = append(reason, "PackValuesUpdated")
			return reason, nil
		},
	}

	d.Set("review_repave_state", "")
	err = validateSystemRepaveApproval(d, m)
	expectedErrMsg := "cluster repave state is pending. \nDue to the following reasons -  \nPackValuesUpdated\nKindly verify the cluster and set `review_repave_state` to `Approved` to continue the repave operation and day 2 operation on the cluster."
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err)
	}

}
