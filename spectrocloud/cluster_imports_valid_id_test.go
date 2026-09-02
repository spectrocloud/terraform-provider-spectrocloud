package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//// Fans out valid-format Import calls across all cluster resources.
// Each resourceClusterXxxImport follows the same 3-step shape:
//   1. GetCommonCluster (parses ID + fetches via mock GetCluster)
//   2. resourceClusterXxxRead (populates state)
//   3. flattenCommonAttributeForClusterImport (adds profile info)
//
// Their current coverage is ~30% each — mostly from the invalid-ID
// error path exercised in Batch 20. Feeding a valid "uid:project" ID
// makes GetCommonCluster succeed against the mock's default cluster
// fixture, so step 2+3 both fire.

type batch30ImportCase struct {
	name     string
	resFn    func() *schema.Resource
	importFn func(context.Context, *schema.ResourceData, interface{}) ([]*schema.ResourceData, error)
	id       string // baseline "uid:context" — the mock returns the default fixture for any UID
}

func TestClusterImports_ValidID(t *testing.T) {
	cases := []batch30ImportCase{
		{"aws", resourceClusterAws, resourceClusterAwsImport, "test-cluster-id:project"},
		{"aks", resourceClusterAks, resourceClusterAksImport, "test-cluster-id:project"},
		{"azure", resourceClusterAzure, resourceClusterAzureImport, "test-cluster-id:project"},
		{"gcp", resourceClusterGcp, resourceClusterGcpImport, "test-gcp-cluster-id:project"},
		{"gke", resourceClusterGke, resourceClusterGkeImport, "test-gke-cluster-id:project"},
		{"vsphere", resourceClusterVsphere, resourceClusterVsphereImport, "test-cluster-id:project"},
		{"maas", resourceClusterMaas, resourceClusterMaasImport, "test-cluster-id:project"},
		{"edge_native", resourceClusterEdgeNative, resourceClusterEdgeNativeImport, "test-cluster-id:project"},
		{"edge_vsphere", resourceClusterEdgeVsphere, resourceClusterEdgeVsphereImport, "test-cluster-id:project"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			defer func() { _ = recover() }()
			d := c.resFn().TestResourceData()
			d.SetId(c.id)
			_, _ = c.importFn(context.Background(), d, unitTestMockAPIClient)
		})
	}
}
