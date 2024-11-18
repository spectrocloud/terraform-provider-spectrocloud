package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertSummaryToIDS(t *testing.T) {
	tests := []struct {
		name     string
		input    []*models.V1UIDSummary
		expected []string
	}{
		{
			name: "Multiple UIDs",
			input: []*models.V1UIDSummary{
				{UID: "uid1"},
				{UID: "uid2"},
				{UID: "uid3"},
			},
			expected: []string{"uid1", "uid2", "uid3"},
		},
		{
			name:     "Empty input",
			input:    []*models.V1UIDSummary{},
			expected: []string(nil),
		},
		{
			name: "Single UID",
			input: []*models.V1UIDSummary{
				{UID: "singleUID"},
			},
			expected: []string{"singleUID"},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []string(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSummaryToIDS(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertToStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []string
	}{
		{
			name:     "All strings",
			input:    []interface{}{"one", "two", "three"},
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "Mixed types",
			input:    []interface{}{"one", 2, "three", 4.0, true},
			expected: []string{"one", "three"},
		},
		{
			name:     "No strings",
			input:    []interface{}{1, 2, 3, 4.5, false},
			expected: []string(nil),
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: []string(nil),
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []string(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToUser(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"first_name": {Type: schema.TypeString, Required: true},
		"last_name":  {Type: schema.TypeString, Required: true},
		"email":      {Type: schema.TypeString, Required: true},
		"team_ids":   {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	}, map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "johndoe@example.com",
		"team_ids":   []interface{}{"team1", "team2"},
	})

	user := toUser(resourceData)

	expectedUser := &models.V1UserEntity{
		Metadata: &models.V1ObjectMeta{
			Name: "John Doe",
		},
		Spec: &models.V1UserSpecEntity{
			EmailID:   "johndoe@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Teams:     []string{"team1", "team2"},
		},
	}

	assert.Equal(t, expectedUser, user)
}

func TestToUserNoTeams(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"first_name": {Type: schema.TypeString, Required: true},
		"last_name":  {Type: schema.TypeString, Required: true},
		"email":      {Type: schema.TypeString, Required: true},
		"team_ids":   {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	}, map[string]interface{}{
		"first_name": "Alice",
		"last_name":  "Smith",
		"email":      "alice@example.com",
	})

	user := toUser(resourceData)

	expectedUser := &models.V1UserEntity{
		Metadata: &models.V1ObjectMeta{
			Name: "Alice Smith",
		},
		Spec: &models.V1UserSpecEntity{
			EmailID:   "alice@example.com",
			FirstName: "Alice",
			LastName:  "Smith",
			Teams:     nil,
		},
	}

	assert.Equal(t, expectedUser, user)
}

func TestSetToStringArray(t *testing.T) {
	// Create a schema.Set with some string values
	input := schema.NewSet(schema.HashString, []interface{}{"id1", "id2", "id3"})

	// Call the function with the set
	result := setToStringArray(input)

	// Define the expected output
	expected := []string{"id1", "id2", "id3"}

	// Assert that the result matches the expected output
	assert.ElementsMatch(t, expected, result)
}

func TestSetToStringArrayEmptySet(t *testing.T) {
	// Create an empty schema.Set
	input := schema.NewSet(schema.HashString, []interface{}{})

	// Call the function with the empty set
	result := setToStringArray(input)

	// Define the expected output for an empty set
	expected := []string{}

	// Assert that the result matches the expected output
	assert.Equal(t, expected, result)
}

func TestToUserWorkspaceRoleMappingEmpty(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"workspace_role": {
			Type: schema.TypeSet,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"workspace": {
						Type: schema.TypeSet,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"id":       {Type: schema.TypeString},
								"role_ids": {Type: schema.TypeSet, Elem: &schema.Schema{Type: schema.TypeString}},
							},
						},
					},
				},
			},
		},
	}, map[string]interface{}{"workspace_role": []interface{}{}})

	result := toUserWorkspaceRoleMapping(d)
	expected := &models.V1WorkspacesRolesPatch{Workspaces: []*models.V1WorkspaceRolesPatch{}}

	assert.Equal(t, expected, result)
}
