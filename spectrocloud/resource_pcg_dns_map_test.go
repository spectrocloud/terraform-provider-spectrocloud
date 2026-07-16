package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// (test file header) CRUD tests for the PCG DNS
// map resource + coverage for the pure helpers.

func preparePCGDNSMapResourceData() *schema.ResourceData {
	d := resourcePrivateCloudGatewayDNSMap().TestResourceData()
	_ = d.Set("private_cloud_gateway_id", "test-pcg-uid")
	_ = d.Set("search_domain_name", "example.com")
	_ = d.Set("data_center", "dc1")
	_ = d.Set("network", "vm-network")
	return d
}

func TestResourcePCGDNSMapCRUD(t *testing.T) {
	testResourceCRUD(t, preparePCGDNSMapResourceData, unitTestMockAPIClient,
		resourcePCGDNSMapCreate, resourcePCGDNSMapRead, resourcePCGDNSMapUpdate, resourcePCGDNSMapDelete)
}

// TestToDNSMap pins the pure schema → model converter in isolation.
func TestToDNSMap(t *testing.T) {
	d := preparePCGDNSMapResourceData()
	got := toDNSMap(d)
	require.NotNil(t, got)
	require.NotNil(t, got.Metadata)
	require.NotNil(t, got.Spec)
	assert.Equal(t, "example.com", got.Metadata.Name)
	require.NotNil(t, got.Spec.DNSName)
	assert.Equal(t, "example.com", *got.Spec.DNSName)
	require.NotNil(t, got.Spec.Datacenter)
	assert.Equal(t, "dc1", *got.Spec.Datacenter)
	require.NotNil(t, got.Spec.Network)
	assert.Equal(t, "vm-network", *got.Spec.Network)
	require.NotNil(t, got.Spec.PrivateGatewayUID)
	assert.Equal(t, "test-pcg-uid", *got.Spec.PrivateGatewayUID)
}

// TestFlattenDNSMap covers the reverse — model → schema.
func TestFlattenDNSMap(t *testing.T) {
	d := resourcePrivateCloudGatewayDNSMap().TestResourceData()
	dc := "dc1"
	name := "example.com"
	network := "vm-network"
	pcg := "test-pcg-uid"
	in := &models.V1VsphereDNSMapping{
		Metadata: &models.V1ObjectMeta{Name: name},
		Spec: &models.V1VsphereDNSMappingSpec{
			Datacenter:        &dc,
			DNSName:           &name,
			Network:           &network,
			PrivateGatewayUID: &pcg,
		},
	}
	require.NoError(t, flattenDNSMap(in, d))
	assert.Equal(t, "example.com", d.Get("search_domain_name"))
	assert.Equal(t, "dc1", d.Get("data_center"))
	assert.Equal(t, "vm-network", d.Get("network"))
	assert.Equal(t, "test-pcg-uid", d.Get("private_cloud_gateway_id"))

	// nil-input branch: helper should return without error and leave state untouched.
	d2 := resourcePrivateCloudGatewayDNSMap().TestResourceData()
	require.NoError(t, flattenDNSMap(nil, d2))
	assert.Empty(t, d2.Get("search_domain_name"))
}
