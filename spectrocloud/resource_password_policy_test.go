package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// ---------------------------------------------------------------------------
// CRUD coverage
//
// Password policy is a singleton resource — no dedicated Create endpoint;
// Create/Update/Delete all POST to the same URL. Read GETs the current
// value. This block adds the mock-driven happy path plus a handful of
// error-path assertions, so the whole resource*.go file goes from ~37%
// to close to full coverage.
// ---------------------------------------------------------------------------

func preparePasswordPolicyResourceData() *schema.ResourceData {
	d := resourcePasswordPolicy().TestResourceData()
	_ = d.Set("password_regex", "")
	_ = d.Set("password_expiry_days", 90)
	_ = d.Set("first_reminder_days", 10)
	_ = d.Set("min_password_length", 8)
	_ = d.Set("min_uppercase_letters", 1)
	_ = d.Set("min_digits", 1)
	_ = d.Set("min_lowercase_letters", 1)
	_ = d.Set("min_special_characters", 1)
	// Read enforces d.Id() == "default-password-policy-id" for flatten to
	// happen — Create sets this ID itself, so we don't preset it in
	// prepare, but a Read-standalone test does need to.
	return d
}

func TestResourcePasswordPolicyCRUD(t *testing.T) {
	testResourceCRUD(t, preparePasswordPolicyResourceData, unitTestMockAPIClient,
		resourcePasswordPolicyCreate, resourcePasswordPolicyRead,
		resourcePasswordPolicyUpdate, resourcePasswordPolicyDelete)
}

func TestResourcePasswordPolicyReadWithoutID(t *testing.T) {
	// The Read handler has a defensive branch: if d.Id() is not the fixed
	// singleton value, it clears d.Id() and returns without flattening.
	// That branch exists so a cross-plane import that stamps the wrong ID
	// doesn't half-populate state — pin it here.
	d := preparePasswordPolicyResourceData()
	d.SetId("some-random-id")
	diags := resourcePasswordPolicyRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Empty(t, d.Id(), "Read should clear a non-canonical ID rather than overwriting state")
}

func TestResourcePasswordPolicyCRUDNegative(t *testing.T) {
	t.Run("Create surfaces API error", func(t *testing.T) {
		testResourceCRUDNegative(t, "Create", preparePasswordPolicyResourceData,
			unitTestMockAPINegativeClient,
			resourcePasswordPolicyCreate, resourcePasswordPolicyRead,
			resourcePasswordPolicyUpdate, resourcePasswordPolicyDelete,
			false, "Invalid password policy")
	})

	// NOTE: negative-path Read is NOT tested here. palette-sdk-go's
	// client.GetPasswordPolicy dereferences resp.Payload before checking
	// err (client/password_policy.go), which panics with SIGSEGV whenever
	// the API returns anything other than 2xx. That's an SDK bug — not
	// something the provider can defend against — and covering it would
	// require dropping the assertion or catching the panic. Flag it as
	// a follow-up rather than pretend the code path is well-behaved.

	t.Run("Update surfaces API error", func(t *testing.T) {
		testResourceCRUDNegative(t, "Update", preparePasswordPolicyResourceData,
			unitTestMockAPINegativeClient,
			resourcePasswordPolicyCreate, resourcePasswordPolicyRead,
			resourcePasswordPolicyUpdate, resourcePasswordPolicyDelete,
			true, "Invalid password policy")
	})

	t.Run("Delete surfaces API error", func(t *testing.T) {
		// Delete's failure surfaces via the same POST /policy endpoint (Delete
		// = revert to defaults, not a DELETE HTTP verb).
		testResourceCRUDNegative(t, "Delete", preparePasswordPolicyResourceData,
			unitTestMockAPINegativeClient,
			resourcePasswordPolicyCreate, resourcePasswordPolicyRead,
			resourcePasswordPolicyUpdate, resourcePasswordPolicyDelete,
			true, "Invalid password policy")
	})
}

// passwordPolicyDiffFixture drives Resource.Diff (which runs the registered
// CustomizeDiff internally) against a brand new resource (nil old state) built
// from cfg. This is the same pattern proven out in
// resource_cluster_profile_test.go's customizeDiffFixture: Resource.Diff builds
// a real *schema.ResourceDiff internally, which a naive &schema.ResourceDiff{}
// literal cannot do since its fields are unexported.
func passwordPolicyDiffFixture(cfg map[string]interface{}) (*terraform.InstanceDiff, error) {
	r := resourcePasswordPolicy()
	return r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(cfg), unitTestMockAPIClient)
}

// TestResourcePasswordPolicyCustomizeDiff pins each branch of the
// CustomizeDiff validator — the "regex + individual mins" conflict, the
// "regex requires expiry" and "regex requires reminder" required-field
// checks, and the passthrough (no regex → no error).
func TestResourcePasswordPolicyCustomizeDiff(t *testing.T) {
	r := resourcePasswordPolicy()
	assert.NotNil(t, r.CustomizeDiff, "CustomizeDiff is wired")

	t.Run("regex conflicts with min_password_length", func(t *testing.T) {
		_, err := passwordPolicyDiffFixture(map[string]interface{}{
			"password_regex":      "^[a-z]+$",
			"min_password_length": 8,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password_regex cannot be used together with min_password_length")
	})

	t.Run("regex requires password_expiry_days", func(t *testing.T) {
		_, err := passwordPolicyDiffFixture(map[string]interface{}{
			"password_regex":       "^[a-z]+$",
			"password_expiry_days": 0,
			"first_reminder_days":  5,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password_expiry_days is required when using password_regex")
	})

	t.Run("regex requires first_reminder_days", func(t *testing.T) {
		_, err := passwordPolicyDiffFixture(map[string]interface{}{
			"password_regex":       "^[a-z]+$",
			"password_expiry_days": 30,
			"first_reminder_days":  0,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "first_reminder_days is required when using password_regex")
	})

	t.Run("regex with all required fields set is valid", func(t *testing.T) {
		diff, err := passwordPolicyDiffFixture(map[string]interface{}{
			"password_regex":       "^[a-z]+$",
			"password_expiry_days": 30,
			"first_reminder_days":  5,
		})
		require.NoError(t, err)
		assert.NotNil(t, diff)
	})

	t.Run("empty regex is a passthrough regardless of min fields", func(t *testing.T) {
		diff, err := passwordPolicyDiffFixture(map[string]interface{}{
			"password_regex":      "",
			"min_password_length": 10,
		})
		require.NoError(t, err)
		assert.NotNil(t, diff)
	})
}

// TestResourcePasswordPolicyImport covers the happy path (matching tenant
// UID resolves) and the mismatch guard (tenant UID differs → error).
func TestResourcePasswordPolicyImport(t *testing.T) {
	t.Run("uid matches tenant", func(t *testing.T) {
		d := resourcePasswordPolicy().TestResourceData()
		d.SetId("test-tenant-uid") // matches getMockUserInfoPayload().tenantUid
		got, err := resourcePasswordPolicyImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "default-password-policy-id", got[0].Id(),
			"import should set the canonical singleton ID")
	})

	t.Run("uid mismatch errors", func(t *testing.T) {
		d := resourcePasswordPolicy().TestResourceData()
		d.SetId("some-other-tenant-uid")
		_, err := resourcePasswordPolicyImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})
}

// TestFlattenPasswordPolicyRegexPreserved ensures the flatten-with-regex
// branch leaves the individual "min_*" fields untouched, which is the
// exact opposite of the without-regex branch. Both paths matter — a
// silent field write on the regex path would cause drift.
func TestFlattenPasswordPolicyRegexPreserved(t *testing.T) {
	d := resourcePasswordPolicy().TestResourceData()
	_ = d.Set("min_password_length", 42)
	require.NoError(t, flattenPasswordPolicy(&models.V1TenantPasswordPolicyEntity{
		Regex:                "^whatever$",
		ExpiryDurationInDays: 30,
		FirstReminderInDays:  3,
	}, d))
	assert.Equal(t, "^whatever$", d.Get("password_regex"))
	assert.Equal(t, 42, d.Get("min_password_length"), "regex flatten must not clobber min_* fields")
}
