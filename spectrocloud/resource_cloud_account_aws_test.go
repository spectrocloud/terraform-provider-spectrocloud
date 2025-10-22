package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestToAwsAccountCTXProjectSecret(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc")
	rd.Set("aws_access_key", "ABCDEFGHIJKLMNOPQRST")
	rd.Set("aws_secret_key", "sasf1424aqsfsdf123423SDFs23412sadf@#$@#$")
	rd.Set("context", "project")
	rd.Set("type", "secret")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("aws_access_key"), acc.Spec.AccessKey)
	assert.Equal(t, rd.Get("aws_secret_key"), acc.Spec.SecretKey)
	assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(*acc.Spec.CredentialType))
}

func TestToAwsAccountCTXTenantSecret(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc")
	rd.Set("aws_access_key", "ABCDEFGHIJKLMNOPQRST")
	rd.Set("aws_secret_key", "sasf1424aqsfsdf123423SDFs23412sadf@#$@#$")
	rd.Set("context", "tenant")
	rd.Set("type", "secret")
	rd.Set("partition", "test_partition")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("aws_access_key"), acc.Spec.AccessKey)
	assert.Equal(t, rd.Get("aws_secret_key"), acc.Spec.SecretKey)
	assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(*acc.Spec.CredentialType))
	assert.Equal(t, rd.Get("partition"), *acc.Spec.Partition)
}

func TestToAwsAccountCTXProjectSecuredAccessKey(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc_secured")
	rd.Set("aws_secured_access_key", "ABCDEFGHIJKLMNOPQRST")
	rd.Set("aws_secret_key", "sasf1424aqsfsdf123423SDFs23412sadf@#$@#$")
	rd.Set("context", "project")
	rd.Set("type", "secret")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("aws_secured_access_key"), acc.Spec.AccessKey)
	assert.Equal(t, rd.Get("aws_secret_key"), acc.Spec.SecretKey)
	assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(*acc.Spec.CredentialType))
}

func TestToAwsAccountCTXTenantSecuredAccessKey(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc_secured")
	rd.Set("aws_secured_access_key", "ABCDEFGHIJKLMNOPQRST")
	rd.Set("aws_secret_key", "sasf1424aqsfsdf123423SDFs23412sadf@#$@#$")
	rd.Set("context", "tenant")
	rd.Set("type", "secret")
	rd.Set("partition", "test_partition")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("aws_secured_access_key"), acc.Spec.AccessKey)
	assert.Equal(t, rd.Get("aws_secret_key"), acc.Spec.SecretKey)
	assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(*acc.Spec.CredentialType))
	assert.Equal(t, rd.Get("partition"), *acc.Spec.Partition)
}

func TestToAwsAccountSecuredAccessKeyPriority(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc_priority")
	// Set secured key - it should take priority even if legacy key exists in state
	rd.Set("aws_secured_access_key", "SECURED_ACCESS_KEY_123")
	rd.Set("aws_secret_key", "sasf1424aqsfsdf123423SDFs23412sadf@#$@#$")
	rd.Set("context", "project")
	rd.Set("type", "secret")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, "SECURED_ACCESS_KEY_123", acc.Spec.AccessKey)
	assert.Equal(t, rd.Get("aws_secret_key"), acc.Spec.SecretKey)
}

func TestToAwsAccountCTXProjectSTS(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc")
	rd.Set("type", "sts")
	rd.Set("arn", "ARN::AWSAD:12312sdTEd")
	rd.Set("external_id", "TEST-External23423ID")
	rd.Set("context", "project")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("arn"), acc.Spec.Sts.Arn)
	assert.Equal(t, rd.Get("external_id"), acc.Spec.Sts.ExternalID)
	assert.Equal(t, "project", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(*acc.Spec.CredentialType))
}

func TestToAwsAccountCTXTenantSTS(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	rd.Set("name", "aws_unit_test_acc")
	rd.Set("type", "sts")
	rd.Set("arn", "ARN::AWSAD:12312sdTEd")
	rd.Set("external_id", "TEST-External23423ID")
	rd.Set("context", "tenant")
	acc, err := toAwsAccount(rd)
	assert.NoError(t, err)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("arn"), acc.Spec.Sts.Arn)
	assert.Equal(t, rd.Get("external_id"), acc.Spec.Sts.ExternalID)
	assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("type"), string(*acc.Spec.CredentialType))
}

func TestFlattenCloudAccountAwsSTS(t *testing.T) {
	// Create a mock ResourceData object
	rd := resourceCloudAccountAws().TestResourceData() // Assuming this method exists

	// Create a mock AWS account model
	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "aws_test_account",
			Annotations: map[string]string{
				"scope": "aws_scope_test",
			},
		},
		Spec: &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSts.Pointer(),
			Sts:            &models.V1AwsStsCredentials{Arn: "test_arn"},
			Partition:      types.Ptr("test_partition"),
			PolicyARNs:     []string{"arn:aws:test_policy1", "arn:aws:test_policy2"},
		},
	}

	// Call the flatten function
	diags, hasError := flattenCloudAccountAws(rd, account)

	// Assertions
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "aws_test_account", rd.Get("name"))
	assert.Equal(t, "aws_scope_test", rd.Get("context"))
	assert.Equal(t, "test_arn", rd.Get("arn"))
	assert.Equal(t, "test_partition", rd.Get("partition"))
	assert.Equal(t, string(models.V1AwsCloudAccountCredentialTypeSts), rd.Get("type"))

	// Handle policy_arns as a *schema.Set
	policyARNs, ok := rd.Get("policy_arns").(*schema.Set)
	if !ok {
		t.Fatalf("Expected policy_arns to be a *schema.Set")
	}

	var actualARNs []string
	for _, v := range policyARNs.List() {
		actualARNs = append(actualARNs, v.(string))
	}

	expectedARNs := []string{"arn:aws:test_policy1", "arn:aws:test_policy2"}
	assert.ElementsMatch(t, expectedARNs, actualARNs)
}

func TestFlattenCloudAccountAws_NonStsType(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()

	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "aws_test_account_secret",
			Annotations: map[string]string{
				"scope": "aws_scope_test_secret",
			},
		},
		Spec: &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
			AccessKey:      "test_access_key_secret",
			Partition:      types.Ptr("test_partition_secret"),
			PolicyARNs:     []string{"arn:aws:test_policy_secret1", "arn:aws:test_policy_secret2"},
		},
	}

	// Call the flatten function
	diags, hasError := flattenCloudAccountAws(rd, account)

	// Assertions
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "aws_test_account_secret", rd.Get("name"))
	assert.Equal(t, "aws_scope_test_secret", rd.Get("context"))
	assert.Equal(t, "test_access_key_secret", rd.Get("aws_access_key"))
	assert.Empty(t, rd.Get("arn")) // Asserting that arn is not set
	assert.Equal(t, "test_partition_secret", rd.Get("partition"))

	// Handle policy_arns as a *schema.Set
	policyARNs, ok := rd.Get("policy_arns").(*schema.Set)
	if !ok {
		t.Fatalf("Expected policy_arns to be a *schema.Set")
	}

	var actualARNs []string
	for _, v := range policyARNs.List() {
		actualARNs = append(actualARNs, v.(string))
	}

	expectedARNs := []string{"arn:aws:test_policy_secret1", "arn:aws:test_policy_secret2"}
	assert.ElementsMatch(t, expectedARNs, actualARNs)
}

func TestFlattenCloudAccountAws_WithSecuredAccessKey(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	// Simulate that aws_secured_access_key was set in the state
	rd.Set("aws_secured_access_key", "existing_secured_key")

	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "aws_test_account_secured",
			Annotations: map[string]string{
				"scope": "aws_scope_test_secured",
			},
		},
		Spec: &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
			AccessKey:      "test_secured_access_key",
			Partition:      types.Ptr("test_partition_secured"),
			PolicyARNs:     []string{"arn:aws:test_policy_secured1"},
		},
	}

	// Call the flatten function
	diags, hasError := flattenCloudAccountAws(rd, account)

	// Assertions
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "aws_test_account_secured", rd.Get("name"))
	assert.Equal(t, "aws_scope_test_secured", rd.Get("context"))
	assert.Equal(t, "test_secured_access_key", rd.Get("aws_secured_access_key"))
	assert.Empty(t, rd.Get("aws_access_key")) // Legacy field should not be set
	assert.Equal(t, "test_partition_secured", rd.Get("partition"))

	// Handle policy_arns as a *schema.Set
	policyARNs, ok := rd.Get("policy_arns").(*schema.Set)
	if !ok {
		t.Fatalf("Expected policy_arns to be a *schema.Set")
	}

	var actualARNs []string
	for _, v := range policyARNs.List() {
		actualARNs = append(actualARNs, v.(string))
	}

	expectedARNs := []string{"arn:aws:test_policy_secured1"}
	assert.ElementsMatch(t, expectedARNs, actualARNs)
}

func TestFlattenCloudAccountAws_LegacyAccessKey(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	// Simulate legacy behavior - aws_secured_access_key is empty/not set

	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "aws_test_account_legacy",
			Annotations: map[string]string{
				"scope": "project",
			},
		},
		Spec: &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
			AccessKey:      "test_legacy_access_key",
			Partition:      types.Ptr("aws"),
		},
	}

	// Call the flatten function
	diags, hasError := flattenCloudAccountAws(rd, account)

	// Assertions
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "aws_test_account_legacy", rd.Get("name"))
	assert.Equal(t, "project", rd.Get("context"))
	assert.Equal(t, "test_legacy_access_key", rd.Get("aws_access_key"))
	assert.Empty(t, rd.Get("aws_secured_access_key")) // Secured field should not be set
	assert.Equal(t, "aws", rd.Get("partition"))
}

func TestFlattenCloudAccountAws_SwitchFromSecuredToLegacy(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	// Simulate scenario where aws_secured_access_key was previously set
	// but now we're reading back an account that should use aws_access_key
	rd.Set("aws_secured_access_key", "old_secured_key")

	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "aws_test_account_switch",
			Annotations: map[string]string{
				"scope": "project",
			},
		},
		Spec: &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
			AccessKey:      "new_access_key",
			Partition:      types.Ptr("aws"),
		},
	}

	// Call the flatten function - it should keep using aws_secured_access_key since it was already set
	diags, hasError := flattenCloudAccountAws(rd, account)

	// Assertions
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "aws_test_account_switch", rd.Get("name"))
	assert.Equal(t, "project", rd.Get("context"))
	assert.Equal(t, "new_access_key", rd.Get("aws_secured_access_key"))
	assert.Empty(t, rd.Get("aws_access_key")) // Legacy field should be cleared to avoid conflicts
	assert.Equal(t, "aws", rd.Get("partition"))
}

func TestFlattenCloudAccountAws_ClearConflictingFieldLegacy(t *testing.T) {
	rd := resourceCloudAccountAws().TestResourceData()
	// Simulate scenario where aws_secured_access_key is NOT set,
	// so aws_access_key should be used and aws_secured_access_key should be cleared

	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: "aws_test_account_clear",
			Annotations: map[string]string{
				"scope": "project",
			},
		},
		Spec: &models.V1AwsCloudAccount{
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
			AccessKey:      "legacy_access_key",
			Partition:      types.Ptr("aws"),
		},
	}

	// Call the flatten function
	diags, hasError := flattenCloudAccountAws(rd, account)

	// Assertions
	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "aws_test_account_clear", rd.Get("name"))
	assert.Equal(t, "project", rd.Get("context"))
	assert.Equal(t, "legacy_access_key", rd.Get("aws_access_key"))
	assert.Empty(t, rd.Get("aws_secured_access_key")) // Should be explicitly cleared to avoid conflicts
	assert.Equal(t, "aws", rd.Get("partition"))
}

func prepareBaseAwsAccountTestData() *schema.ResourceData {
	d := resourceCloudAccountAws().TestResourceData()
	d.SetId("test-aws-account-1")
	_ = d.Set("name", "test-aws-account")
	_ = d.Set("context", "project")
	_ = d.Set("aws_access_key", "test-access-key")
	_ = d.Set("aws_secret_key", "test-secret-key")
	_ = d.Set("type", "secret")
	_ = d.Set("arn", "test-arn")
	_ = d.Set("external_id", "test-external-id")
	_ = d.Set("partition", "aws")
	_ = d.Set("policy_arns", []string{"test-policy-arn"})
	return d
}

func prepareSecuredAwsAccountTestData() *schema.ResourceData {
	d := resourceCloudAccountAws().TestResourceData()
	d.SetId("test-aws-account-1")
	_ = d.Set("name", "test-aws-account-secured")
	_ = d.Set("context", "project")
	_ = d.Set("aws_secured_access_key", "test-secured-access-key")
	_ = d.Set("aws_secret_key", "test-secret-key")
	_ = d.Set("type", "secret")
	_ = d.Set("partition", "aws")
	_ = d.Set("policy_arns", []string{"test-policy-arn"})
	return d
}

func TestResourceCloudAccountAwsCreate(t *testing.T) {
	ctx := context.Background()
	d := prepareBaseAwsAccountTestData()
	diags := resourceCloudAccountAwsCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsRead(t *testing.T) {
	ctx := context.Background()
	d := prepareBaseAwsAccountTestData()
	diags := resourceCloudAccountAwsRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}
func TestResourceCloudAccountAwsUpdate(t *testing.T) {
	ctx := context.Background()
	d := prepareBaseAwsAccountTestData()
	diags := resourceCloudAccountAwsUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}
func TestResourceCloudAccountAwsDelete(t *testing.T) {
	ctx := context.Background()
	d := prepareBaseAwsAccountTestData()
	diags := resourceCloudAccountAwsDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}

func TestResourceCloudAccountAwsImport(t *testing.T) {
	ctx := context.Background()
	d := prepareBaseAwsAccountTestData()
	d.SetId("test-import-acc-id:project")
	_, err := resourceAccountAwsImport(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, err)
	assert.Equal(t, "test-import-acc-id", d.Id())
}

func TestResourceCloudAccountAwsCreateWithSecuredAccessKey(t *testing.T) {
	ctx := context.Background()
	d := prepareSecuredAwsAccountTestData()
	diags := resourceCloudAccountAwsCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsReadWithSecuredAccessKey(t *testing.T) {
	ctx := context.Background()
	d := prepareSecuredAwsAccountTestData()
	diags := resourceCloudAccountAwsRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsUpdateWithSecuredAccessKey(t *testing.T) {
	ctx := context.Background()
	d := prepareSecuredAwsAccountTestData()
	diags := resourceCloudAccountAwsUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAwsDeleteWithSecuredAccessKey(t *testing.T) {
	ctx := context.Background()
	d := prepareSecuredAwsAccountTestData()
	diags := resourceCloudAccountAwsDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}
