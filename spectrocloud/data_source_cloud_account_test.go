package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

type cloudAccountReadTestCase struct {
	name                   string
	prepareData            func() *schema.ResourceData
	readFunc               func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	setupAttrs             map[string]interface{} // name or id, plus optional context/cloud
	useNegativeClient      bool
	expectedErrorSubstring string // only for negative case
}

func TestDataSourceCloudAccountRead_TableDriven(t *testing.T) {
	ctx := context.Background()

	// Define configs per cloud type (nameSlug for test data, prepare + read + error message).
	configs := []struct {
		name     string
		nameSlug string // lowercase for "test-<slug>-account-1" to match mock
		prepare  func() *schema.ResourceData
		read     func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
		errorMsg string
	}{
		{"AWS", "aws", dataSourceCloudAccountAws().TestResourceData, dataSourceCloudAccountAwsRead, "Unable to find aws cloud account"},
		{"Azure", "azure", dataSourceCloudAccountAzure().TestResourceData, dataSourceCloudAccountAzureRead, "Unable to find azure cloud account"},
		{"GCP", "gcp", dataSourceCloudAccountGcp().TestResourceData, dataSourceCloudAccountGcpRead, "Unable to find gcp cloud account"},
		{"Vsphere", "vsphere", dataSourceCloudAccountVsphere().TestResourceData, dataSourceCloudAccountVsphereRead, "Unable to find vsphere cloud account"},
		{"Openstack", "openstack", dataSourceCloudAccountOpenStack().TestResourceData, dataSourceCloudAccountOpenStackRead, "Unable to find openstack cloud account"},
		{"Maas", "maas", dataSourceCloudAccountMaas().TestResourceData, dataSourceCloudAccountMaasRead, "Unable to find maas cloud account"},
		{"Custom", "custom", dataSourceCloudAccountCustom().TestResourceData, dataSourceCloudAccountCustomRead, "Unable to find cloud account"},
	}

	var testCases []cloudAccountReadTestCase
	for _, c := range configs {
		// ReadByName: set name (AWS also needs context)
		attrsByName := map[string]interface{}{"name": "test-" + c.nameSlug + "-account-1"}
		if c.name == "AWS" {
			attrsByName["context"] = "project"
		}
		if c.name == "Custom" {
			attrsByName["cloud"] = "nutanix"
		}
		testCases = append(testCases, cloudAccountReadTestCase{
			name:              c.name + "_ReadByName",
			prepareData:       c.prepare,
			readFunc:          c.read,
			setupAttrs:        attrsByName,
			useNegativeClient: false,
		})

		// ReadByID: set id (Custom also needs cloud)
		attrsByID := map[string]interface{}{"id": "test-" + c.nameSlug + "-account-id-1"}
		if c.name == "Custom" {
			attrsByID["cloud"] = "nutanix"
		}
		testCases = append(testCases, cloudAccountReadTestCase{
			name:              c.name + "_ReadByID",
			prepareData:       c.prepare,
			readFunc:          c.read,
			setupAttrs:        attrsByID,
			useNegativeClient: false,
		})

		// ReadNegative: set name, use negative client
		attrsNeg := map[string]interface{}{"name": "test-" + c.nameSlug + "-account-1"}
		if c.name == "Custom" {
			attrsNeg["cloud"] = "nutanix"
		}
		testCases = append(testCases, cloudAccountReadTestCase{
			name:                   c.name + "_ReadNegative",
			prepareData:            c.prepare,
			readFunc:               c.read,
			setupAttrs:             attrsNeg,
			useNegativeClient:      true,
			expectedErrorSubstring: c.errorMsg,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := tc.prepareData()
			for k, v := range tc.setupAttrs {
				_ = d.Set(k, v)
			}

			meta := unitTestMockAPIClient
			if tc.useNegativeClient {
				meta = unitTestMockAPINegativeClient
			}

			diags := tc.readFunc(ctx, d, meta)

			if tc.useNegativeClient {
				assertFirstDiagMessage(t, diags, tc.expectedErrorSubstring)
			} else {
				assert.Empty(t, diags, "expected no diagnostics")
			}
		})
	}
}
