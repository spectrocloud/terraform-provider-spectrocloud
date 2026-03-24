package virtualmachine

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	"github.com/stretchr/testify/assert"
	k8sv1 "k8s.io/api/core/v1"
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
					Type:   kubevirtapiv1.VirtualMachineConditionType("InvalidType"),
					Status: k8sv1.ConditionStatus("InvalidStatus"),
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
		input    []kubevirtapiv1.VirtualMachineCondition
		expected []interface{}
	}{
		{
			name: "valid input",
			input: []kubevirtapiv1.VirtualMachineCondition{
				{
					Type:    kubevirtapiv1.VirtualMachineConditionType("Ready"),
					Status:  k8sv1.ConditionStatus("True"),
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
			input:    []kubevirtapiv1.VirtualMachineCondition{},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenVirtualMachineConditions(tt.input)
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
		input    []kubevirtapiv1.VirtualMachineStateChangeRequest
		expected []interface{}
	}{
		{
			name:     "empty input",
			input:    []kubevirtapiv1.VirtualMachineStateChangeRequest{},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenVirtualMachineStateChangeRequests(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandVirtualMachineStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected kubevirtapiv1.VirtualMachineStatus
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
			expected: kubevirtapiv1.VirtualMachineStatus{
				Created: true,
				Ready:   false,
				Conditions: []kubevirtapiv1.VirtualMachineCondition{
					{
						Type:   kubevirtapiv1.VirtualMachineConditionType("Ready"),
						Status: k8sv1.ConditionStatus("True"),
					},
				},
				StateChangeRequests: []kubevirtapiv1.VirtualMachineStateChangeRequest{
					{
						Action: kubevirtapiv1.StateChangeRequestAction("Start"),
						Data:   utils.ExpandStringMap(map[string]interface{}{"key1": "value1"}),
					},
				},
			},
		},
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: kubevirtapiv1.VirtualMachineStatus{},
		},
		{
			name: "partial input",
			input: []interface{}{
				map[string]interface{}{
					"created": true,
				},
			},
			expected: kubevirtapiv1.VirtualMachineStatus{
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
		input    kubevirtapiv1.VirtualMachineStatus
		expected []interface{}
	}{
		{
			name: "full input",
			input: kubevirtapiv1.VirtualMachineStatus{
				Created: true,
				Ready:   false,
				Conditions: []kubevirtapiv1.VirtualMachineCondition{
					{
						Type:   kubevirtapiv1.VirtualMachineConditionType("Ready"),
						Status: k8sv1.ConditionStatus("True"),
					},
				},
				StateChangeRequests: []kubevirtapiv1.VirtualMachineStateChangeRequest{
					{
						Action: kubevirtapiv1.StateChangeRequestAction("Start"),
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
			input: kubevirtapiv1.VirtualMachineStatus{},
			expected: []interface{}{
				map[string]interface{}{
					"created":               false,
					"ready":                 false,
					"conditions":            []interface{}{},
					"state_change_requests": []interface{}{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattenVirtualMachineStatus(tt.input)
		})
	}
}
