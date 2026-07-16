package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		hasErrors bool
	}{
		{name: "empty", timezone: "", hasErrors: false},
		{name: "utc", timezone: "UTC", hasErrors: false},
		{name: "iana timezone", timezone: "Asia/Kolkata", hasErrors: false},
		{name: "missing slash", timezone: "AsiaKolkata", hasErrors: true},
		{name: "contains space", timezone: "Asia/ Kolkata", hasErrors: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateTimezone(tt.timezone, "cluster_timezone")
			if tt.hasErrors {
				assert.NotEmpty(t, errs)
				return
			}
			assert.Empty(t, errs)
		})
	}
}

func TestToTagsMapAndFlattenTagsMap(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"tags_map": {
			Type:     schema.TypeMap,
			Optional: true,
		},
	}, map[string]interface{}{
		"tags_map": map[string]interface{}{
			"owner": "platform",
			"cost":  "",
		},
	})

	tags := toTagsMap(d)
	assert.Equal(t, "platform", tags["owner"])
	assert.Equal(t, "spectro__tag", tags["cost"])

	flattened := flattenTagsMap(tags)
	assert.Equal(t, "platform", flattened["owner"])
	assert.Equal(t, "spectro__tag", flattened["cost"])
}

func TestFlattenTagsMapEmpty(t *testing.T) {
	assert.Nil(t, flattenTagsMap(map[string]string{}))
}
