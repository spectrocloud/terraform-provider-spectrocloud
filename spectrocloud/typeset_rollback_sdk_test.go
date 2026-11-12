package spectrocloud

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// These tests run the real Terraform CLI against a minimal provider to pin down the SDK behaviour
// cluster_profile rollback depends on (PLT-2508): on a failed apply, terraform-plugin-sdk replaces
// a TypeSet value the provider set with the planned value whenever the planned value is wholly
// known. A Computed-only attribute in the set element makes a changed element's planned value
// unknown, so the provider's rollback survives.

func typeSetRollbackTestProvider(withComputedName, failUpdate bool) *schema.Provider {
	elem := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {Type: schema.TypeString, Required: true},
		},
	}
	if withComputedName {
		elem.Schema["name"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	}

	read := func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
		if !withComputedName {
			return nil
		}
		// Mirrors the provider filling cluster_profile.name from Palette on read.
		profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
		for _, p := range profiles {
			m := p.(map[string]interface{})
			m["name"] = "name-of-" + m["id"].(string)
		}
		return diag.FromErr(d.Set("cluster_profile", profiles))
	}

	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"rollbacktest_cluster": {
				Schema: map[string]*schema.Schema{
					"cluster_profile": {
						Type:     schema.TypeSet,
						Optional: true,
						Set:      schema.HashResource(elem),
						Elem:     elem,
					},
				},
				CreateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
					d.SetId("cluster-1")
					return read(ctx, d, m)
				},
				ReadContext: read,
				DeleteContext: func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
					return nil
				},
				UpdateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
					if failUpdate {
						oldProfile, _ := d.GetChange("cluster_profile")
						_ = d.Set("cluster_profile", oldProfile)
						return diag.Errorf("profile update rejected")
					}
					return read(ctx, d, m)
				},
			},
		},
	}
}

func typeSetRollbackTestConfig(profileID string) string {
	return `
resource "rollbacktest_cluster" "c" {
  cluster_profile {
    id = "` + profileID + `"
  }
}
`
}

func typeSetRollbackTestFactories(withComputedName, failUpdate bool) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"rollbacktest": func() (*schema.Provider, error) {
			return typeSetRollbackTestProvider(withComputedName, failUpdate), nil
		},
	}
}

func TestTypeSetRollback_WhollyKnownPlanOverwritesRollback(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: typeSetRollbackTestFactories(false, true),
		Steps: []resource.TestStep{
			{Config: typeSetRollbackTestConfig("profile-a")},
			{Config: typeSetRollbackTestConfig("profile-b"), ExpectError: regexp.MustCompile(`profile update rejected`)},
			// Empty plan against profile-b: the SDK discarded the rollback and saved the rejected value.
			{Config: typeSetRollbackTestConfig("profile-b"), PlanOnly: true},
		},
	})
}

func TestTypeSetRollback_ComputedElementFieldKeepsRollback(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: typeSetRollbackTestFactories(true, true),
		Steps: []resource.TestStep{
			{Config: typeSetRollbackTestConfig("profile-a")},
			{Config: typeSetRollbackTestConfig("profile-b"), ExpectError: regexp.MustCompile(`profile update rejected`)},
			// Empty plan against profile-a: the rollback survived the failed apply.
			{Config: typeSetRollbackTestConfig("profile-a"), PlanOnly: true},
			// And the rejected change is still pending, so it can be retried.
			{Config: typeSetRollbackTestConfig("profile-b"), PlanOnly: true, ExpectNonEmptyPlan: true},
		},
	})
}

func TestTypeSetRollback_ComputedElementFieldSuccessfulUpdate(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: typeSetRollbackTestFactories(true, false),
		Steps: []resource.TestStep{
			{Config: typeSetRollbackTestConfig("profile-a")},
			{
				Config: typeSetRollbackTestConfig("profile-b"),
				Check:  resource.TestCheckTypeSetElemNestedAttrs("rollbacktest_cluster.c", "cluster_profile.*", map[string]string{"id": "profile-b", "name": "name-of-profile-b"}),
			},
			// A refreshed computed name must not show as drift.
			{Config: typeSetRollbackTestConfig("profile-b"), PlanOnly: true},
		},
	})
}
