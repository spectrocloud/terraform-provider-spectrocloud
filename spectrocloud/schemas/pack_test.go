package schemas

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestPackSchema_ValuesAttributeOptional(t *testing.T) {
	// Get the pack schema
	packSchema := PackSchema()

	// Get the "values" attribute schema
	valuesSchema := packSchema.Elem.(*schema.Resource).Schema["values"]

	// Check if the "Optional" field is set to true
	assert.True(t, valuesSchema.Optional, "The 'values' attribute should be optional")
}
