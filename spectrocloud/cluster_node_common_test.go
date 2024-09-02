package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

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
