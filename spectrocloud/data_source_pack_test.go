package spectrocloud

import (
	"context"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourcePackRead(t *testing.T) {

	m := &client.V1Client{
		GetPacksFn: func(filters []string, regUid string) ([]*models.V1PackSummary, error) {
			var p []*models.V1PackSummary
			return p, nil
		},
		GetHelmRegistryFn: func(string) (*models.V1HelmRegistry, error) {
			hr := models.V1HelmRegistry{
				APIVersion: "",
				Kind:       "",
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "testRegistry",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "",
				},
				Spec: &models.V1HelmRegistrySpec{
					Auth:        nil,
					Endpoint:    nil,
					IsPrivate:   true,
					Name:        "testRegistry",
					RegistryUID: "test-reg-uid",
					Scope:       "",
				},
				Status: nil,
			}

			return &hr, nil
		},
	}
	// Initialize the test schema.ResourceData
	resourceData := dataSourcePack().TestResourceData()
	// Context
	ctx := context.TODO()
	// Test case: type is "manifest"
	err := resourceData.Set("type", "manifest")
	if err != nil {
		return
	}
	diags := dataSourcePackRead(ctx, resourceData, m)
	assert.Empty(t, diags)
	// Test case: type is "helm" and registry is private
	err = resourceData.Set("type", "helm")
	if err != nil {
		return
	}
	err = resourceData.Set("registry_uid", "test-registry")
	if err != nil {
		return
	}
	diags = dataSourcePackRead(ctx, resourceData, m)
	assert.Empty(t, diags)
	// Test case: type is "oci" and registry_uid is set
	err = resourceData.Set("type", "oci")
	if err != nil {
		return
	}
	diags = dataSourcePackRead(ctx, resourceData, m)
	assert.Empty(t, diags)

	// Test case: GetPacks returns an error
	resourceData.Set("type", "some-other-type")
	diags = dataSourcePackRead(ctx, resourceData, m)
	assert.NotEmpty(t, diags)
	assert.Equal(t, diag.Error, diags[0].Severity)

	err = resourceData.Set("type", "")
	if err != nil {
		return
	}
	diags = dataSourcePackRead(ctx, resourceData, m)
	assert.NotEmpty(t, diags)
	assert.Equal(t, diag.Error, diags[0].Severity)

	err = resourceData.Set("type", "")
	if err != nil {
		return
	}
	err = resourceData.Set("name", "centos-aws")
	if err != nil {
		return
	}
	err = resourceData.Set("version", "7.7")
	if err != nil {
		return
	}
	diags = dataSourcePackRead(ctx, resourceData, m)
	assert.NotEmpty(t, diags)
	assert.Equal(t, diag.Error, diags[0].Severity)

	m2 := &client.V1Client{
		GetPacksFn: func(filters []string, regUid string) ([]*models.V1PackSummary, error) {
			var p []*models.V1PackSummary
			pack := &models.V1PackSummary{
				APIVersion: "",
				Kind:       "",
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "centos-aws",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "testpackuid",
				},
				Spec: &models.V1PackSummarySpec{
					AddonSubType: "",
					AddonType:    "",
					Annotations:  nil,
					CloudTypes:   []string{"aws"},
					Digest:       "",
					DisplayName:  "Centos",
					Eol:          "",
					Group:        "",
					Layer:        "",
					LogoURL:      "",
					Manifests:    nil,
					Name:         "centos-aws",
					Presets:      nil,
					RegistryUID:  "",
					Schema:       nil,
					Template:     nil,
					Type:         "",
					Values:       "",
					Version:      "",
				},
				Status: nil,
			}
			p = append(p, pack)
			return p, nil
		},
	}
	err = resourceData.Set("type", "")
	if err != nil {
		return
	}
	err = resourceData.Set("name", "centos-aws")
	if err != nil {
		return
	}
	err = resourceData.Set("version", "7.7")
	if err != nil {
		return
	}
	diags = dataSourcePackRead(ctx, resourceData, m2)
	assert.Empty(t, diags)

	m2 = &client.V1Client{
		GetPacksFn: func(filters []string, regUid string) ([]*models.V1PackSummary, error) {
			var p []*models.V1PackSummary
			pack := &models.V1PackSummary{
				Metadata: &models.V1ObjectMeta{
					Name: "centos-aws",
					UID:  "testpackuid",
				},
				Spec: &models.V1PackSummarySpec{
					CloudTypes:  []string{"aws"},
					DisplayName: "Centos",
					Name:        "centos-aws",
				},
				Status: nil,
			}
			pack1 := &models.V1PackSummary{
				Metadata: &models.V1ObjectMeta{
					Name: "centos-aws-1",
					UID:  "testpackuid-1",
				},
				Spec: &models.V1PackSummarySpec{
					CloudTypes:  []string{"aws"},
					DisplayName: "Centos-1",
					Name:        "centos-aws-1",
				},
				Status: nil,
			}
			p = append(p, pack)
			p = append(p, pack1)
			return p, nil
		},
	}
	err = resourceData.Set("type", "")
	if err != nil {
		return
	}
	err = resourceData.Set("name", "centos-aws")
	if err != nil {
		return
	}
	err = resourceData.Set("version", "7.7")
	if err != nil {
		return
	}
	diags = dataSourcePackRead(ctx, resourceData, m2)
	assert.NotEmpty(t, diags)
	assert.Equal(t, diag.Error, diags[0].Severity)
}
