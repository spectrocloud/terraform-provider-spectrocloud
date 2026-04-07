package virtualmachine

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	"github.com/stretchr/testify/assert"
	// kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func TestExpandVirtualMachineConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []*models.V1VMVirtualMachineCondition
		wantErr  bool
	}{
		{
			name: "valid input",
			input: []interface{}{
				map[string]interface{}{
					"type":    "Ready",
					"status":  "True",
					"reason":  "Initialized",
					"message": "VM is ready",
				},
			},
			expected: []*models.V1VMVirtualMachineCondition{
				{
					Type:    utils.PtrToString("Ready"),
					Status:  utils.PtrToString("True"),
					Reason:  "Initialized",
					Message: "VM is ready",
				},
			},
			wantErr: false,
		},
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: []*models.V1VMVirtualMachineCondition{},
			wantErr:  false,
		},
		{
			name: "invalid input",
			input: []interface{}{
				map[string]interface{}{
					"type":   "InvalidType",
					"status": "InvalidStatus",
				},
			},
			expected: []*models.V1VMVirtualMachineCondition{
				{
					Type:   utils.PtrToString("InvalidType"),
					Status: utils.PtrToString("InvalidStatus"),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandVirtualMachineConditions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandVirtualMachineConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenVirtualMachineConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    []*models.V1VMVirtualMachineCondition
		expected []interface{}
	}{
		{
			name: "valid input",
			input: []*models.V1VMVirtualMachineCondition{
				{
					Type:    utils.PtrToString("Ready"),
					Status:  utils.PtrToString("True"),
					Reason:  "Initialized",
					Message: "VM is ready",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"type":    "Ready",
					"status":  "True",
					"reason":  "Initialized",
					"message": "VM is ready",
				},
			},
		},
		{
			name:     "empty input",
			input:    []*models.V1VMVirtualMachineCondition{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenVirtualMachineConditionsFromVM(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandVirtualMachineStateChangeRequests(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []*models.V1VMVirtualMachineStateChangeRequest
	}{
		{
			name: "valid input",
			input: []interface{}{
				map[string]interface{}{
					"action": "Start",
					"data": map[string]interface{}{
						"key1": "value1",
					},
					"uid": "1234",
				},
			},
			expected: []*models.V1VMVirtualMachineStateChangeRequest{
				{
					Action: utils.PtrToString("Start"),
					Data:   utils.ExpandStringMap(map[string]interface{}{"key1": "value1"}),
					UID:    "1234",
				},
			},
		},
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: []*models.V1VMVirtualMachineStateChangeRequest{},
		},
		{
			name: "partial input",
			input: []interface{}{
				map[string]interface{}{
					"action": "Stop",
				},
			},
			expected: []*models.V1VMVirtualMachineStateChangeRequest{
				{
					Action: utils.PtrToString("Stop"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandVirtualMachineStateChangeRequests(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenVirtualMachineStateChangeRequests(t *testing.T) {
	tests := []struct {
		name     string
		input    []*models.V1VMVirtualMachineStateChangeRequest
		expected []interface{}
	}{
		{
			name:     "empty input",
			input:    []*models.V1VMVirtualMachineStateChangeRequest{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenVirtualMachineStateChangeRequestsFromVM(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandVirtualMachineStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected models.V1ClusterVirtualMachineStatus
	}{
		{
			name: "full input",
			input: []interface{}{
				map[string]interface{}{
					"created": true,
					"ready":   false,
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Ready",
							"status": "True",
						},
					},
					"state_change_requests": []interface{}{
						map[string]interface{}{
							"action": "Start",
							"data": map[string]interface{}{
								"key1": "value1",
							},
							"uid": "1234",
						},
					},
				},
			},
			expected: models.V1ClusterVirtualMachineStatus{
				Created: true,
				Ready:   false,
				Conditions: []*models.V1VMVirtualMachineCondition{
					{
						Type:   utils.PtrToString("Ready"),
						Status: utils.PtrToString("True"),
					},
				},
				StateChangeRequests: []*models.V1VMVirtualMachineStateChangeRequest{
					{
						Action: utils.PtrToString("Start"),
						Data:   utils.ExpandStringMap(map[string]interface{}{"key1": "value1"}),
					},
				},
			},
		},
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: models.V1ClusterVirtualMachineStatus{},
		},
		{
			name: "partial input",
			input: []interface{}{
				map[string]interface{}{
					"created": true,
				},
			},
			expected: models.V1ClusterVirtualMachineStatus{
				Created: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := expandVirtualMachineStatus(tt.input)
			assert.NoError(t, err)
		})
	}
}

func TestFlattenVirtualMachineStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.V1ClusterVirtualMachineStatus
		expected []interface{}
	}{
		{
			name: "full input",
			input: &models.V1ClusterVirtualMachineStatus{
				Created: true,
				Ready:   false,
				Conditions: []*models.V1VMVirtualMachineCondition{
					{
						Type:   utils.PtrToString("Ready"),
						Status: utils.PtrToString("True"),
					},
				},
				StateChangeRequests: []*models.V1VMVirtualMachineStateChangeRequest{
					{
						Action: utils.PtrToString("Start"),
						Data:   utils.ExpandStringMap(map[string]interface{}{"key1": "value1"}),
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"created": true,
					"ready":   false,
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Ready",
							"status": "True",
						},
					},
					"state_change_requests": []interface{}{
						map[string]interface{}{
							"action": "Start",
							"data": map[string]interface{}{
								"key1": "value1",
							},
							"uid": "1234",
						},
					},
				},
			},
		},
		{
			name:  "empty input",
			input: &models.V1ClusterVirtualMachineStatus{},
			expected: []interface{}{
				map[string]interface{}{
					"created":               false,
					"ready":                 false,
					"conditions":            nil,
					"state_change_requests": nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattenVirtualMachineStatusFromVM(tt.input)
		})
	}
}
