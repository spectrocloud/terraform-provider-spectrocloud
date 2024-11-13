package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToRole(t *testing.T) {

	d := resourceRole().TestResourceData()
	err := d.Set("name", "test-role")
	if err != nil {
		return
	}
	err = d.Set("type", "project")
	if err != nil {
		return
	}
	err = d.Set("permissions", []interface{}{"bbb"})
	if err != nil {
		return
	}

	role := toRole(d)

	expected := &models.V1Role{
		Metadata: &models.V1ObjectMeta{
			Annotations: map[string]string{
				"scope": "project",
			},
			Name: "test-role",
		},
		Spec: &models.V1RoleSpec{
			Permissions: []string{"bbb"},
			Scope:       models.V1Scope("project"),
			Type:        "user",
		},
		Status: &models.V1RoleStatus{
			IsEnabled: true,
		},
	}

	assert.Equal(t, expected, role)
}

func TestFlattenRole(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"type": {
			Type:     schema.TypeString,
			Required: true,
		},
		"permissions": {
			Type:     schema.TypeSet,
			Required: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}, map[string]interface{}{})

	role := &models.V1Role{
		Metadata: &models.V1ObjectMeta{
			Name: "test-role",
		},
		Spec: &models.V1RoleSpec{
			Permissions: []string{"read", "write"},
			Scope:       models.V1Scope("admin"),
		},
	}

	err := flattenRole(d, role)
	assert.NoError(t, err)
	assert.Equal(t, "test-role", d.Get("name"))
	assert.Equal(t, "admin", d.Get("type"))
	assert.ElementsMatch(t, []interface{}{"read", "write"}, d.Get("permissions").(*schema.Set).List())
}
