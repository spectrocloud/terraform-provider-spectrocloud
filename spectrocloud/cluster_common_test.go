package spectrocloud

import (
	"github.com/spectrocloud/hapi/models"
	"reflect"
	"testing"
)

func TestToAdditionalNodePoolLabels(t *testing.T) {
	// Test case 1: When 'additional_labels' is nil, the function should return an empty map
	result := toAdditionalNodePoolLabels(map[string]interface{}{"additional_labels": nil})
	if len(result) != 0 {
		t.Errorf("Expected an empty map, got %v", result)
	}

	// Test case 2: When 'additional_labels' is an empty map, the function should return an empty map
	result = toAdditionalNodePoolLabels(map[string]interface{}{"additional_labels": map[string]interface{}{}})
	if len(result) != 0 {
		t.Errorf("Expected an empty map, got %v", result)
	}

	// Test case 3: When 'additional_labels' contains valid data, the function should return the expected result
	input := map[string]interface{}{
		"additional_labels": map[string]interface{}{
			"label1": "value1",
			"label2": "value2",
		},
	}
	expected := map[string]string{
		"label1": "value1",
		"label2": "value2",
	}

	result = toAdditionalNodePoolLabels(input)
	if len(result) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	for key, value := range expected {
		if result[key] != value {
			t.Errorf("Expected %s=%s, got %s=%s", key, value, key, result[key])
		}
	}
}

func TestToClusterTaints(t *testing.T) {
	// Test case 1: When 'taints' is nil, the function should return an empty slice
	result := toClusterTaints(map[string]interface{}{"taints": nil})
	if len(result) != 0 {
		t.Errorf("Expected an empty slice, got %v", result)
	}

	// Test case 2: When 'taints' is an empty slice, the function should return an empty slice
	result = toClusterTaints(map[string]interface{}{"taints": []interface{}{}})
	if len(result) != 0 {
		t.Errorf("Expected an empty slice, got %v", result)
	}

	// Test case 3: When 'taints' contains valid data, the function should return the expected result
	input := map[string]interface{}{
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
	}

	expected := []*models.V1Taint{
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
	}

	result = toClusterTaints(input)
	if len(result) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	for i, expectedTaint := range expected {
		if *result[i] != *expectedTaint {
			t.Errorf("Expected %v, got %v", expectedTaint, result[i])
		}
	}
}

func TestToClusterTaint(t *testing.T) {
	// Test case 1: When 'clusterTaint' is a valid map, the function should return the expected result
	input := map[string]interface{}{
		"key":    "key1",
		"value":  "value1",
		"effect": "NoSchedule",
	}

	expected := &models.V1Taint{
		Key:    "key1",
		Value:  "value1",
		Effect: "NoSchedule",
	}

	result := toClusterTaint(input)
	if *result != *expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFlattenClusterTaints(t *testing.T) {
	// Test case 1: When 'items' is an empty slice, the function should return an empty slice
	result := flattenClusterTaints([]*models.V1Taint{})
	if len(result) != 0 {
		t.Errorf("Expected an empty slice, got %v", result)
	}

	// Test case 2: When 'items' contains valid taints, the function should return the expected result
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

	expected := []interface{}{
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
	}

	result = flattenClusterTaints([]*models.V1Taint{taint1, taint2})
	if len(result) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	for i, expectedTaint := range expected {
		if !reflect.DeepEqual(result[i], expectedTaint) {
			t.Errorf("Expected %v, got %v", expectedTaint, result[i])
		}
	}
}

func TestFlattenAdditionalLabelsAndTaints(t *testing.T) {
	// Test case 1: When 'labels' is empty and 'intaints' is empty, oi should be updated accordingly
	oi := make(map[string]interface{})
	FlattenAdditionalLabelsAndTaints(make(map[string]string), []*models.V1Taint{}, oi)

	expectedOi := map[string]interface{}{
		"additional_labels": map[string]interface{}{},
	}

	if !reflect.DeepEqual(oi, expectedOi) {
		t.Errorf("Expected %v, got %v", expectedOi, oi)
	}

	// Test case 2: When 'labels' is not empty, 'additional_labels' in oi should be updated
	labels := map[string]string{
		"label1": "value1",
		"label2": "value2",
	}
	oi = make(map[string]interface{})
	FlattenAdditionalLabelsAndTaints(labels, []*models.V1Taint{}, oi)

	expectedOi = map[string]interface{}{
		"additional_labels": map[string]interface{}{
			"label1": "value1",
			"label2": "value2",
		},
	}

	if !mapsAreEqual(oi, expectedOi) {
		t.Errorf("Expected %v, got %v", expectedOi, oi)
	}

	// Test case 3: When 'intaints' is not empty, 'taints' in oi should be updated
	taints := []*models.V1Taint{
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
	}
	oi = make(map[string]interface{})
	FlattenAdditionalLabelsAndTaints(labels, taints, oi)
	//var v1 interface{} = "value1"
	//var v2 interface{} = "value2"
	expectedOi = map[string]interface{}{
		"additional_labels": map[string]interface{}{
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
	}

	if !mapsAreEqual(oi, expectedOi) {
		t.Errorf("Expected %v, got %v", expectedOi, oi)
	}
}

func mapsAreEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for key, _ := range a {
		_, ok := b[key]
		if !ok {
			return false
		}

	}

	return true
}
