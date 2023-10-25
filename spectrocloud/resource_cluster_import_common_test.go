package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestReadCommonAttributes(t *testing.T) {
	tests := []struct {
		name            string
		inputAttributes map[string]interface{}
		expectedResults map[string]interface{}
		expectError     bool
	}{
		{
			name:            "default values",
			inputAttributes: map[string]interface{}{
				// No input attributes to test default scenario
			},
			expectedResults: map[string]interface{}{
				"force_delete":       false,
				"force_delete_delay": 20,
				"os_patch_on_boot":   false,
				"skip_completion":    false,
				"apply_setting":      "DownloadAndInstall",
			},
		},
		{
			name: "non-default values",
			inputAttributes: map[string]interface{}{
				"force_delete":       true,
				"force_delete_delay": 30,
				"os_patch_on_boot":   true,
				"skip_completion":    true,
				"apply_setting":      "CustomValue",
			},
			expectedResults: map[string]interface{}{
				"force_delete":       true,
				"force_delete_delay": 30,
				"os_patch_on_boot":   true,
				"skip_completion":    true,
				"apply_setting":      "CustomValue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"force_delete":       {Type: schema.TypeBool},
				"force_delete_delay": {Type: schema.TypeInt},
				"os_patch_on_boot":   {Type: schema.TypeBool},
				"skip_completion":    {Type: schema.TypeBool},
				"apply_setting":      {Type: schema.TypeString},
			}, tt.inputAttributes)

			err := ReadCommonAttributes(rd)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for key, expectedValue := range tt.expectedResults {
				assert.Equal(t, expectedValue, rd.Get(key))
			}
		})
	}
}
