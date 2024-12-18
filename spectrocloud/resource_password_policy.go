package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"time"
)

func resourcePasswordPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePasswordPolicyCreate,
		ReadContext:   resourcePasswordPolicyRead,
		UpdateContext: resourcePasswordPolicyUpdate,
		DeleteContext: resourcePasswordPolicyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"password_regex": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ConflictsWith: []string{"min_password_length", "min_uppercase_letters",
					"min_digits", "min_lowercase_letters", "min_special_characters"},
				RequiredWith: []string{"password_expiry_days", "first_reminder_days"},
				Description:  "A regular expression (regex) to define custom password patterns, such as enforcing specific characters or sequences in the password.",
			},
			"password_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      999,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "The number of days before the password expires. Must be between 1 and 1000 days. Defines how often passwords must be changed.",
			},
			"first_reminder_days": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "The number of days before the password expiry to send the first reminder to the user. Default is 5 days before expiry.",
			},
			"min_password_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     12,
				Description: "The minimum length required for the password. Enforces a stronger password policy by ensuring a minimum number of characters.",
			},
			"min_uppercase_letters": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The minimum number of uppercase letters (A-Z) required in the password. Helps ensure password complexity with a mix of case-sensitive characters.",
			},
			"min_digits": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The minimum number of numeric digits (0-9) required in the password. Ensures that passwords contain numerical characters.",
			},
			"min_lowercase_letters": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The minimum number of lowercase letters (a-z) required in the password. Ensures that lowercase characters are included for password complexity.",
			},
			"min_special_characters": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The minimum number of special characters (e.g., !, @, #, $, %) required in the password. This increases the password's security level by including symbols.",
			},
		},
	}
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
	err = d.Set("password_regex", passwordPolicy.Regex)
	if err != nil {
		return err
	}
	err = d.Set("password_expiry_days", passwordPolicy.ExpiryDurationInDays)
	if err != nil {
		return err
	}
	err = d.Set("first_reminder_days", passwordPolicy.FirstReminderInDays)
	if err != nil {
		return err
	}
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
	//c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
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
