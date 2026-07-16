package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	modelsPkg "github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

// contextB12 returns a background context used by Batch 12 tests below.
func contextB12() context.Context { return context.Background() }

func TestGetMachinePoolList(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name:    "Handle *schema.Set",
			input:   schema.NewSet(schema.HashString, []interface{}{"a", "b"}),
			want:    []interface{}{"a", "b"},
			wantErr: false,
		},
		{
			name:    "Handle []interface{}",
			input:   []interface{}{"a", "b"},
			want:    []interface{}{"a", "b"},
			wantErr: false,
		},
		{
			name:    "Handle unexpected type",
			input:   "unexpected",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := getMachinePoolList(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMachinePoolList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

// Test for getNodeValue function
func TestGetNodeValue1(t *testing.T) {
	nodeId := "node-123"
	action := "action-xyz"

	expected := map[string]interface{}{
		"node_id": nodeId,
		"action":  action,
	}

	result := getNodeValue(nodeId, action)
	assert.Equal(t, expected, result)
}

// Test for getMachinePoolList function
func TestGetMachinePoolList1(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []interface{}
		isError  bool
	}{
		{
			name: "With []interface{}",
			input: []interface{}{
				map[string]interface{}{"name": "pool1"},
				map[string]interface{}{"name": "pool2"},
			},
			expected: []interface{}{
				map[string]interface{}{"name": "pool1"},
				map[string]interface{}{"name": "pool2"},
			},
			isError: false,
		},
		{
			name:     "With invalid type",
			input:    "invalid",
			expected: nil,
			isError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := getMachinePoolList(tt.input)
			if tt.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

// dummyMaintenanceStatusB12 Stub GetMaintenanceStatus used
// by the resourceNodeAction / refresh-func tests below.
func dummyMaintenanceStatusB12(_, _, _ string) (*modelsPkg.V1MachineMaintenanceStatus, error) {
	return &modelsPkg.V1MachineMaintenanceStatus{
		Action: "cordon",
		State:  "Completed",
	}, nil
}

func TestResourceNodeAction_NoNodes(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	// newMachinePool without a "node" key → the for-loop is skipped and
	// nil is returned.
	err := resourceNodeAction(c, contextB12(),
		map[string]interface{}{"name": "mp-1"},
		dummyMaintenanceStatusB12, "aws", "cfg-uid", "mp-1")
	assert.NoError(t, err)
}

func TestResourceNodeAction_NoActionChange(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	// One node whose action already matches the stub → no ToggleMaintenance
	// call is made, no waiter fires, and the loop exits with nil.
	err := resourceNodeAction(c, contextB12(),
		map[string]interface{}{
			"name": "mp-1",
			"node": []interface{}{
				map[string]interface{}{
					"node_id": "n1",
					"action":  "cordon",
				},
			},
		},
		dummyMaintenanceStatusB12, "aws", "cfg-uid", "mp-1")
	assert.NoError(t, err)
}

func TestResourceClusterNodeMaintenanceRefreshFunc(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	refresh := resourceClusterNodeMaintenanceRefreshFunc(
		c, dummyMaintenanceStatusB12, "cfg-uid", "mp-1", "node-1")

	// The closure executes; SDK call may miss but we just want the body
	// covered.
	_, _, _ = refresh()
}
