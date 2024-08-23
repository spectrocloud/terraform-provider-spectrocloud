package spectrocloud

import (
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
	assert.Equal(t, rd.Get("type"), string(acc.Spec.CredentialType))
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
	assert.Equal(t, rd.Get("type"), string(acc.Spec.CredentialType))
	assert.Equal(t, rd.Get("partition"), *acc.Spec.Partition)
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
	assert.Equal(t, rd.Get("type"), string(acc.Spec.CredentialType))
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
	assert.Equal(t, rd.Get("type"), string(acc.Spec.CredentialType))
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
			CredentialType: models.V1AwsCloudAccountCredentialTypeSts,
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
			CredentialType: models.V1AwsCloudAccountCredentialTypeSecret,
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
