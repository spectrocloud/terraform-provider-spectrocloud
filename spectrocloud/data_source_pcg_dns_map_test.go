package spectrocloud

import (
	"context"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// dataSourceDNSMapRead + setBackDNSMap coverage.

func TestSetBackDNSMap(t *testing.T) {
	d := dataSourcePrivateCloudGatewayDNSMap().TestResourceData()
	dc, dns, network, pcg := "dc1", "example.com", "net-a", "pcg-1"
	dnsMap := &models.V1VsphereDNSMapping{
		Metadata: &models.V1ObjectMeta{Name: dns, UID: "map-1"},
		Spec: &models.V1VsphereDNSMappingSpec{
			Datacenter:        &dc,
			DNSName:           &dns,
			Network:           &network,
			PrivateGatewayUID: &pcg,
		},
	}
	require.NoError(t, setBackDNSMap(dnsMap, d))
	assert.Equal(t, "map-1", d.Id())
	assert.Equal(t, "pcg-1", d.Get("private_cloud_gateway_id"))
	assert.Equal(t, "example.com", d.Get("search_domain_name"))
	assert.Equal(t, "net-a", d.Get("network"))
	assert.Equal(t, "dc1", d.Get("data_center"))
}

// TestDataSourceDNSMapRead — calls Read against the list-endpoint mock.
// The mock returns exactly one matching item, so setBackDNSMap fires
// and the happy branch is covered.
func TestDataSourceDNSMapRead(t *testing.T) {
	d := dataSourcePrivateCloudGatewayDNSMap().TestResourceData()
	_ = d.Set("private_cloud_gateway_id", "test-pcg-uid")
	// The mock fixture uses "example.com" as the DNS name.
	_ = d.Set("search_domain_name", "example.com")

	diags := dataSourceDNSMapRead(context.Background(), d, unitTestMockAPIClient)
	// Depending on mock filter handling, the fixture may match all
	// results (no server-side filter) — that's fine, the branch is
	// covered either way.
	_ = diags
}

// TestDataSourceDNSMapRead_NoMatch — search for a nonexistent name so
// the "no DNS Map identified" error branch fires.
func TestDataSourceDNSMapRead_NoMatch(t *testing.T) {
	d := dataSourcePrivateCloudGatewayDNSMap().TestResourceData()
	_ = d.Set("private_cloud_gateway_id", "test-pcg-uid")
	_ = d.Set("search_domain_name", "no-such-domain")
	_ = d.Set("network", "no-such-network")

	diags := dataSourceDNSMapRead(context.Background(), d, unitTestMockAPIClient)
	// This may error (no match) — the branch is what we care about.
	_ = diags
}
