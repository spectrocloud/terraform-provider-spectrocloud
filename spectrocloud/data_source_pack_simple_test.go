package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseDataSourceSimplePackResourceData() *schema.ResourceData {
	d := dataSourcePackSimple().TestResourceData()
	_ = d.Set("name", "k8")
	return d
}

func TestDataSourceSimplePacksReadManifest(t *testing.T) {
	d := prepareBaseDataSourceSimplePackResourceData()
	_ = d.Set("type", "manifest")
	diags := dataSourcePackReadSimple(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestDataSourceSimplePacksReadManifestWithoutReg(t *testing.T) {
	d := prepareBaseDataSourceSimplePackResourceData()
	_ = d.Set("type", "other")
	diags := dataSourcePackReadSimple(context.Background(), d, unitTestMockAPIClient)
	assertFirstDiagMessage(t, diags, "No registry uid provided.")
}

func TestDataSourceSimplePacksRead(t *testing.T) {
	d := prepareBaseDataSourceSimplePackResourceData()
	_ = d.Set("type", "other")
	_ = d.Set("registry_uid", "test-reg-uid")
	_ = d.Set("version", "1.0")
	diags := dataSourcePackReadSimple(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestDataSourceSimplePacksReadNoPackFound(t *testing.T) {
	d := prepareBaseDataSourceSimplePackResourceData()
	_ = d.Set("type", "other")
	_ = d.Set("registry_uid", "test-reg-uid")
	_ = d.Set("version", "1.0")
	diags := dataSourcePackReadSimple(context.Background(), d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "No values for pack found.")
}
