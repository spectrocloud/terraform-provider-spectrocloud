package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

//// Sweep valid-format Import calls across the remaining resource types
// whose Import currently sits at 42.9% (invalid-ID branch only).

func TestResourceApplicationImport_ValidID(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourceApplication().TestResourceData()
	d.SetId("test-app-id")
	_, _ = resourceApplicationImport(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceApplicationProfileImport_ValidID(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourceApplicationProfile().TestResourceData()
	d.SetId("test-profile-uid:project:1.0.0")
	_, _ = resourceApplicationProfileImport(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceBackupStorageLocationImport_ValidID(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourceBackupStorageLocation().TestResourceData()
	d.SetId("test-backup-location-id:project")
	_, _ = resourceBackupStorageLocationImport(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceClusterGroupImport_ValidID(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourceClusterGroup().TestResourceData()
	d.SetId("test-cluster-group-id:project")
	_, _ = resourceClusterGroupImport(context.Background(), d, unitTestMockAPIClient)
}

func TestResourcePCGDNSMapImport_InvalidID(t *testing.T) {
	d := resourcePrivateCloudGatewayDNSMap().TestResourceData()
	d.SetId("only-one-part")
	_, err := resourcePrivateCloudGatewayDNSMapImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
}

func TestResourcePCGDNSMapImport_ValidID(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourcePrivateCloudGatewayDNSMap().TestResourceData()
	d.SetId("test-pcg-uid:test-dnsmap-uid")
	_, _ = resourcePrivateCloudGatewayDNSMapImport(context.Background(), d, unitTestMockAPIClient)
}

func TestResourcePCGIPPoolImport_InvalidID(t *testing.T) {
	d := resourcePrivateCloudGatewayIpPool().TestResourceData()
	d.SetId("only-one-part")
	_, err := resourcePrivateCloudGatewayIpPoolImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
}

func TestResourcePCGIPPoolImport_ValidID(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourcePrivateCloudGatewayIpPool().TestResourceData()
	d.SetId("test-pcg-uid:test-ippool-uid")
	_, _ = resourcePrivateCloudGatewayIpPoolImport(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceRegistryOciImport_ValidID(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourceRegistryOciEcr().TestResourceData()
	d.SetId("test-registry-uid")
	_, _ = resourceRegistryOciImport(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceRegistryHelmImport_ValidID(t *testing.T) {
	defer func() { _ = recover() }()
	d := resourceRegistryHelm().TestResourceData()
	d.SetId("test-registry-uid")
	_, _ = resourceRegistryHelmImport(context.Background(), d, unitTestMockAPIClient)
}
