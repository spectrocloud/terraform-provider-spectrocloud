package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// PLT-2483: Hubble (PEM-11828) exposes seven `{{.spectro.system.oidc.*}}`
// system macros that a pack's `values` YAML can reference so Hubble injects
// the cluster's OIDC issuer/client config at the keys and format each
// distribution expects (e.g. CK8s needs `--`-prefixed
// extraKubeAPIServerArgs keys). The provider must never resolve these
// macros itself and must never see the resolved client secret -- Hubble
// resolves them downstream of the Palette API this provider talks to.
//
// oidcMacroPackValues mirrors the CK8s example from the ticket: all seven
// macros, including the `--`-prefixed keys that must not be touched by any
// provider-side normalization.
const oidcMacroPackValues = `extraKubeAPIServerArgs:
  --oidc-issuer-url: "{{.spectro.system.oidc.issuerurl}}"
  --oidc-client-id: "{{.spectro.system.oidc.clientid}}"
  --oidc-client-secret: "{{.spectro.system.oidc.clientsecret}}"
  --oidc-username-claim: "{{.spectro.system.oidc.usernameclaim}}"
  --oidc-groups-claim: "{{.spectro.system.oidc.groupsclaim}}"
  --oidc-signing-algs: "{{.spectro.system.oidc.extrascopes}}"
  --oidc-idp-config-name: "{{.spectro.system.oidc.idpconfigname}}"
`

// TestToPack_OIDCSystemMacros_PassThroughVerbatim guards the write path
// (day-0 create and day-2 pack update share toPack/setPackValues): the
// macros must reach the API request exactly as authored, with no provider
// templating or escaping of the `{{...}}`/`--`-prefixed-key syntax.
func TestToPack_OIDCSystemMacros_PassThroughVerbatim(t *testing.T) {
	pSrc := map[string]interface{}{
		"name":   "ck8s",
		"values": oidcMacroPackValues,
		"tag":    "1.0.0",
		"type":   "spectro",
	}

	result := toPack(nil, pSrc)

	require.NotNil(t, result)
	assert.Equal(t, oidcMacroPackValues, result.Values, "pack values containing OIDC system macros must not be altered on write")
}

// TestFlattenPacksWithRegistryMaps_OIDCSystemMacros_PassThroughVerbatim
// guards the read path for spectrocloud_cluster_profile: a pack whose
// Values the API returns verbatim (the real behavior -- Hubble resolves
// macros at reconcile time, not inside the Palette core API this flatten
// reads from) must land in state unchanged, never masked or rewritten.
func TestFlattenPacksWithRegistryMaps_OIDCSystemMacros_PassThroughVerbatim(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	packs := []*models.V1PackRef{
		{
			Name:   types.Ptr("ck8s"),
			Tag:    "1.0.0",
			Type:   "spectro",
			Values: oidcMacroPackValues,
		},
	}

	result, err := flattenPacksWithRegistryMaps(c, nil, packs, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, result, 1)

	pack := result[0].(map[string]interface{})
	assert.Equal(t, oidcMacroPackValues, pack["values"], "pack values containing OIDC system macros must not be altered on read")
}

// TestEnrichClusterProfilesWithPacks_OIDCSystemMacros_PerClusterOverride
// covers the other read path every spectrocloud_cluster_<cloud> resource
// shares: a pack value override attached directly to a cluster (not a
// profile) must also pass the macros through unchanged.
func TestEnrichClusterProfilesWithPacks_OIDCSystemMacros_PerClusterOverride(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID: "profile-with-ck8s",
					Packs: []*models.V1PackRef{
						{Name: types.Ptr("ck8s"), PackUID: "pack-uid-ck8s", Type: "spectro", Values: oidcMacroPackValues},
					},
				},
			},
		},
	}

	d := resourceClusterEks().TestResourceData()
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "profile-with-ck8s",
			"pack": []interface{}{
				map[string]interface{}{"name": "ck8s"},
			},
		},
	}))

	clusterProfiles := []interface{}{
		map[string]interface{}{"id": "profile-with-ck8s", "pack": []interface{}{}},
	}

	result, err := enrichClusterProfilesWithPacks(c, d, cluster, clusterProfiles)
	require.NoError(t, err)
	require.Len(t, result, 1)

	packs := result[0].(map[string]interface{})["pack"].([]interface{})
	require.Len(t, packs, 1)
	assert.Equal(t, oidcMacroPackValues, packs[0].(map[string]interface{})["values"], "per-cluster pack override values containing OIDC system macros must not be altered on read")
}

// TestResourceClusterProfileCustomizeDiff_OIDCMacrosNoOpRefresh_UnderFlag
// guards the immutable-clusterprofiles CustomizeDiff guardrail
// (hasSemanticPackChange/canonicalizePackList): a plain refresh where state
// carries the API-normalized form of a macro-bearing values string (a
// trailing newline) and config has the same content without it must not be
// treated as a real edit. This is the exact false-positive class documented
// on hasSemanticPackChange, reproduced with OIDC macro content specifically
// since `{{`/`}}`/`--`-prefixed keys are unusual enough to warrant their own
// regression guard.
func TestResourceClusterProfileCustomizeDiff_OIDCMacrosNoOpRefresh_UnderFlag(t *testing.T) {
	orig := ProviderFeaturePreview
	defer func() { ProviderFeaturePreview = orig }()
	ProviderFeaturePreview = map[string]bool{"immutable-clusterprofiles": true}

	_, err := customizeDiffPackFixture(
		oidcMacroPackValues,
		oidcMacroPackValues[:len(oidcMacroPackValues)-1], // config form has no trailing newline
		1,
	)
	assert.NoError(t, err, "a no-op refresh of OIDC-macro pack values must not trip the immutability guardrail")
}

// TestResourceClusterProfileCustomizeDiff_OIDCMacrosRealEdit_UnderFlag is
// the symmetric guard: a genuine content edit alongside the macros (adding
// a new extraKubeAPIServerArgs key) must still trip the guardrail -- the
// macro syntax must not accidentally make the semantic-change comparison
// blind to real edits.
func TestResourceClusterProfileCustomizeDiff_OIDCMacrosRealEdit_UnderFlag(t *testing.T) {
	orig := ProviderFeaturePreview
	defer func() { ProviderFeaturePreview = orig }()
	ProviderFeaturePreview = map[string]bool{"immutable-clusterprofiles": true}

	edited := oidcMacroPackValues + "  --oidc-required-claim: \"email=verified\"\n"
	_, err := customizeDiffPackFixture(oidcMacroPackValues, edited, 1)
	assert.Error(t, err, "a real edit alongside OIDC macros must still trip the immutability guardrail")
	assert.Contains(t, err.Error(), "pack")
}
