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
