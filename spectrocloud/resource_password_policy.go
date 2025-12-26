package spectrocloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourcePasswordPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePasswordPolicyCreate,
		ReadContext:   resourcePasswordPolicyRead,
		UpdateContext: resourcePasswordPolicyUpdate,
		DeleteContext: resourcePasswordPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePasswordPolicyImport,
		},
		CustomizeDiff: resourcePasswordPolicyCustomizeDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"password_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A regular expression (regex) to define custom password patterns, such as enforcing specific characters or sequences in the password.",
			},
			"password_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      999,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "The number of days before the password expires. Must be between 1 and 1000 days. Defines how often passwords must be changed.  Default is `999` days for expiry. Conflicts with `min_password_length`, `min_uppercase_letters`, `min_digits`, `min_lowercase_letters`, `min_special_characters`",
			},
			"first_reminder_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "The number of days before the password expiry to send the first reminder to the user. Default is `5` days before expiry.",
			},
			"min_password_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum length required for the password. Enforces a stronger password policy by ensuring a minimum number of characters.  Default minimum length is `6`.",
			},
			"min_uppercase_letters": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum number of uppercase letters (A-Z) required in the password. Helps ensure password complexity with a mix of case-sensitive characters. Minimum length of upper case should be `1`.",
			},
			"min_digits": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum number of numeric digits (0-9) required in the password. Ensures that passwords contain numerical characters. Minimum length of digit should be `1`.",
			},
			"min_lowercase_letters": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum number of lowercase letters (a-z) required in the password. Ensures that lowercase characters are included for password complexity. Minimum length of lower case should be `1`.",
			},
			"min_special_characters": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The minimum number of special characters (e.g., !, @, #, $, %) required in the password. This increases the password's security level by including symbols. Minimum special characters should be `1`.",
			},
		},
	}
}

func resourcePasswordPolicyCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
	passwordRegex := diff.Get("password_regex").(string)

	// If password_regex is set, check that the individual password requirements are not set
	if passwordRegex != "" {
		conflictingFields := []string{
			"min_password_length",
			"min_uppercase_letters",
			"min_digits",
			"min_lowercase_letters",
			"min_special_characters",
		}

		for _, field := range conflictingFields {
			if val := diff.Get(field); val != nil && val != 0 {
				return fmt.Errorf("password_regex cannot be used together with %s. Use either password_regex for custom patterns or the individual minimum requirements", field)
			}
		}

		// When using password_regex, password_expiry_days and first_reminder_days are required
		if diff.Get("password_expiry_days").(int) == 0 {
			return fmt.Errorf("password_expiry_days is required when using password_regex")
		}
		if diff.Get("first_reminder_days").(int) == 0 {
			return fmt.Errorf("first_reminder_days is required when using password_regex")
		}
	}

	return nil
}

func toPasswordPolicy(d *schema.ResourceData) (*models.V1TenantPasswordPolicyEntity, error) {
	if d.Get("password_regex").(string) != "" {
		return &models.V1TenantPasswordPolicyEntity{
			IsRegex:              true,
			Regex:                d.Get("password_regex").(string),
			ExpiryDurationInDays: int64(d.Get("password_expiry_days").(int)),
			FirstReminderInDays:  int64(d.Get("first_reminder_days").(int)),
		}, nil
	}
	return &models.V1TenantPasswordPolicyEntity{
		ExpiryDurationInDays:      int64(d.Get("password_expiry_days").(int)),
		FirstReminderInDays:       int64(d.Get("first_reminder_days").(int)),
		IsRegex:                   false,
		MinLength:                 int64(d.Get("min_password_length").(int)),
		MinNumOfBlockLetters:      int64(d.Get("min_uppercase_letters").(int)),
		MinNumOfDigits:            int64(d.Get("min_digits").(int)),
		MinNumOfSmallLetters:      int64(d.Get("min_lowercase_letters").(int)),
		MinNumOfSpecialCharacters: int64(d.Get("min_special_characters").(int)),
		Regex:                     "",
	}, nil
}

func toPasswordPolicyDefault(d *schema.ResourceData) (*models.V1TenantPasswordPolicyEntity, error) {
	return &models.V1TenantPasswordPolicyEntity{
		ExpiryDurationInDays:      999,
		FirstReminderInDays:       5,
		IsRegex:                   false,
		MinLength:                 6,
		MinNumOfBlockLetters:      1,
		MinNumOfDigits:            1,
		MinNumOfSmallLetters:      1,
		MinNumOfSpecialCharacters: 1,
		Regex:                     "",
	}, nil
}

func flattenPasswordPolicy(passwordPolicy *models.V1TenantPasswordPolicyEntity, d *schema.ResourceData) error {
	var err error
	if passwordPolicy.Regex != "" {
		err = d.Set("password_regex", passwordPolicy.Regex)
		if err != nil {
			return err
		}
	} else {
		err = d.Set("min_password_length", passwordPolicy.MinLength)
		if err != nil {
			return err
		}
		err = d.Set("min_uppercase_letters", passwordPolicy.MinNumOfBlockLetters)
		if err != nil {
			return err
		}
		err = d.Set("min_digits", passwordPolicy.MinNumOfDigits)
		if err != nil {
			return err
		}
		err = d.Set("min_lowercase_letters", passwordPolicy.MinNumOfSmallLetters)
		if err != nil {
			return err
		}
		err = d.Set("min_special_characters", passwordPolicy.MinNumOfSpecialCharacters)
		if err != nil {
			return err
		}
	}
	err = d.Set("password_expiry_days", passwordPolicy.ExpiryDurationInDays)
	if err != nil {
		return err
	}
	err = d.Set("first_reminder_days", passwordPolicy.FirstReminderInDays)
	if err != nil {
		return err
	}

	return nil
}

func resourcePasswordPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	passwordPolicy, err := toPasswordPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	// For Password Policy we don't have support for creation it's always an update
	err = c.UpdatePasswordPolicy(tenantUID, passwordPolicy)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("default-password-policy-id")
	return diags
}

func resourcePasswordPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return handleReadError(d, err, diags)
	}
	resp, err := c.GetPasswordPolicy(tenantUID)
	if err != nil {
		return handleReadError(d, err, diags)
	}
	// handling case for cross-plane for singleton resource
	if d.Id() != "default-password-policy-id" {
		// If we are not reading the default password policy, we should not set the ID
		d.SetId("")
		return diags
	}
	err = flattenPasswordPolicy(resp, d)
	if err != nil {
		return nil
	}
	return diags
}

func resourcePasswordPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	passwordPolicy, err := toPasswordPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	// For Password Policy we don't have support for creation it's always an update
	err = c.UpdatePasswordPolicy(tenantUID, passwordPolicy)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourcePasswordPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	// We can't delete the base password policy, instead
	passwordPolicy, err := toPasswordPolicyDefault(d)
	if err != nil {
		return diag.FromErr(err)
	}
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	// For Password Policy we don't have support for creation it's always an update
	err = c.UpdatePasswordPolicy(tenantUID, passwordPolicy)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

func resourcePasswordPolicyImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	givenTenantId := d.Id()
	actualTenantId, err := c.GetTenantUID()
	if err != nil {
		return nil, err
	}
	if givenTenantId != actualTenantId {
		return nil, fmt.Errorf("tenant id is not valid with current user: %v", diags)
	}
	diags = resourcePasswordPolicyRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read password policy for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}
