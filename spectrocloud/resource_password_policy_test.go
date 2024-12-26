package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestToPasswordPolicy(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"password_regex": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"password_expiry_days": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"first_reminder_days": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"min_password_length": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"min_uppercase_letters": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"min_digits": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"min_lowercase_letters": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"min_special_characters": {
			Type:     schema.TypeInt,
			Optional: true,
		},
	}

	testCases := []struct {
		name        string
		input       map[string]interface{}
		expected    *models.V1TenantPasswordPolicyEntity
		expectError bool
	}{
		{
			name: "Password regex defined",
			input: map[string]interface{}{
				"password_regex":       "^(?=.*[A-Z])(?=.*[a-z])(?=.*\\d).+$",
				"password_expiry_days": 90,
				"first_reminder_days":  10,
			},
			expected: &models.V1TenantPasswordPolicyEntity{
				IsRegex:              true,
				Regex:                "^(?=.*[A-Z])(?=.*[a-z])(?=.*\\d).+$",
				ExpiryDurationInDays: 90,
				FirstReminderInDays:  10,
			},
			expectError: false,
		},
		{
			name: "No regex, full policy specified",
			input: map[string]interface{}{
				"password_regex":         "",
				"password_expiry_days":   90,
				"first_reminder_days":    10,
				"min_password_length":    12,
				"min_uppercase_letters":  2,
				"min_digits":             3,
				"min_lowercase_letters":  4,
				"min_special_characters": 1,
			},
			expected: &models.V1TenantPasswordPolicyEntity{
				IsRegex:                   false,
				Regex:                     "",
				ExpiryDurationInDays:      90,
				FirstReminderInDays:       10,
				MinLength:                 12,
				MinNumOfBlockLetters:      2,
				MinNumOfDigits:            3,
				MinNumOfSmallLetters:      4,
				MinNumOfSpecialCharacters: 1,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, resourceSchema, tc.input)
			result, err := toPasswordPolicy(resourceData)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestToPasswordPolicyDefault(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{}

	resourceData := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	result, err := toPasswordPolicyDefault(resourceData)

	assert.NoError(t, err)
	expected := &models.V1TenantPasswordPolicyEntity{
		ExpiryDurationInDays:      999,
		FirstReminderInDays:       5,
		IsRegex:                   false,
		MinLength:                 6,
		MinNumOfBlockLetters:      1,
		MinNumOfDigits:            1,
		MinNumOfSmallLetters:      1,
		MinNumOfSpecialCharacters: 1,
		Regex:                     "",
	}
	assert.Equal(t, expected, result)
}

func TestFlattenPasswordPolicy(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"password_regex":         {Type: schema.TypeString, Optional: true},
		"password_expiry_days":   {Type: schema.TypeInt, Optional: true},
		"first_reminder_days":    {Type: schema.TypeInt, Optional: true},
		"min_password_length":    {Type: schema.TypeInt, Optional: true},
		"min_uppercase_letters":  {Type: schema.TypeInt, Optional: true},
		"min_digits":             {Type: schema.TypeInt, Optional: true},
		"min_lowercase_letters":  {Type: schema.TypeInt, Optional: true},
		"min_special_characters": {Type: schema.TypeInt, Optional: true},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})

	t.Run("with regex", func(t *testing.T) {
		passwordPolicy := &models.V1TenantPasswordPolicyEntity{
			Regex:                "^[a-zA-Z0-9]+$",
			ExpiryDurationInDays: 90,
			FirstReminderInDays:  10,
		}

		err := flattenPasswordPolicy(passwordPolicy, resourceData)
		assert.NoError(t, err)

		assert.Equal(t, "^[a-zA-Z0-9]+$", resourceData.Get("password_regex"))
		assert.Equal(t, 90, resourceData.Get("password_expiry_days"))
		assert.Equal(t, 10, resourceData.Get("first_reminder_days"))
	})

	t.Run("without regex", func(t *testing.T) {
		passwordPolicy := &models.V1TenantPasswordPolicyEntity{
			ExpiryDurationInDays:      90,
			FirstReminderInDays:       10,
			MinLength:                 8,
			MinNumOfBlockLetters:      2,
			MinNumOfDigits:            2,
			MinNumOfSmallLetters:      2,
			MinNumOfSpecialCharacters: 1,
			Regex:                     "",
		}
		err := resourceData.Set("password_regex", "")
		if err != nil {
			return
		}
		err = flattenPasswordPolicy(passwordPolicy, resourceData)
		assert.NoError(t, err)

		assert.Equal(t, "", resourceData.Get("password_regex"))
		assert.Equal(t, 90, resourceData.Get("password_expiry_days"))
		assert.Equal(t, 10, resourceData.Get("first_reminder_days"))
		assert.Equal(t, 8, resourceData.Get("min_password_length"))
		assert.Equal(t, 2, resourceData.Get("min_uppercase_letters"))
		assert.Equal(t, 2, resourceData.Get("min_digits"))
		assert.Equal(t, 2, resourceData.Get("min_lowercase_letters"))
		assert.Equal(t, 1, resourceData.Get("min_special_characters"))
	})
}
