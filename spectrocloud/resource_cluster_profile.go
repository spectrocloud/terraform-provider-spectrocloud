package spectrocloud

import (
	"context"
	"fmt"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterProfileCreate,
		ReadContext:   resourceClusterProfileRead,
		UpdateContext: resourceClusterProfileUpdate,
		DeleteContext: resourceClusterProfileDelete,
		CustomizeDiff: resourceClusterProfileCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterProfileImport,
		},
		Description: "The Cluster Profile resource allows you to create and manage cluster profiles.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Second),
			Update: schema.DefaultTimeout(20 * time.Second),
			Delete: schema.DefaultTimeout(20 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the cluster profile.",
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.0.0", // default as in UI
				Description: "Version of the cluster profile. Defaults to '1.0.0'. " +
					"\n\n" +
					"Default behavior (no feature flag set): changing this value on an existing " +
					"profile updates the version in place via `PUT /v1/clusterprofiles/{uid}`, " +
					"which destroys the previous version. This is the legacy behavior preserved " +
					"for backward compatibility. " +
					"\n\n" +
					"When the `immutable-clusterprofiles` feature_preview flag is enabled, " +
					"changing this value triggers a Terraform resource **replacement** " +
					"(`ForceNew`) instead of an in-place update. This is the standard Terraform " +
					"Plugin SDK v2 pattern for immutable-versioned resources. Combined with " +
					"`skip_destroy = true` and `lifecycle { create_before_destroy = true }`, " +
					"the new version is created by cloning from the existing Palette lineage " +
					"while the previous version is preserved untouched in Palette. The Terraform " +
					"resource id is set once at Create time and never mutates mid-update, so it " +
					"respects the SDK v2 contract that a resource's primary id is stable across " +
					"in-place updates -- outputs that reference `.id` always reflect the current " +
					"version without needing `terraform apply -refresh-only`.",
			},
			"skip_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "When `true`, `terraform destroy` removes the cluster profile from " +
					"Terraform state without calling the Palette delete API, leaving the " +
					"underlying profile version intact in Palette. " +
					"\n\n" +
					"This is the standard Terraform Plugin SDK v2 preservation pattern for " +
					"immutable-versioned resources. Combined with the `immutable-clusterprofiles` " +
					"feature_preview flag and `lifecycle { create_before_destroy = true }`, it " +
					"lets you bump the `version` field as a normal in-HCL edit while every " +
					"previous version stays preserved in Palette -- Terraform's state advances " +
					"cleanly to the new version while older versions remain immutable in Palette. " +
					"Defaults to `false`.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant", "system"}, false),
				Description: "The context of the cluster profile. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the cluster profile.",
			},
			"cloud": {
				Type:     schema.TypeString,
				Default:  "all",
				Optional: true,
				// Removing validation to support custom clouds
				// ValidateFunc: validation.StringInSlice([]string{"all", "aws", "azure", "gcp", "vsphere", "maas", "virtual", "baremetal", "eks", "aks", "edge", "edge-native", "generic", "gke"}, false),
				ForceNew: true,
				Description: "Specify the infrastructure provider the cluster profile is for. Only Palette supported infrastructure providers can be used. The supported cloud types are - `all, aws, azure, gcp, vsphere, maas, virtual, baremetal, eks, aks, edge-native, generic, and gke` or any custom cloud provider registered in Palette, e.g., `nutanix`." +
					"If the value is set to `all`, then the type must be set to `add-on`. Otherwise, the cluster profile may be incompatible with other providers. Default value is `all`.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "add-on",
				ValidateFunc: validation.StringInSlice([]string{"add-on", "cluster", "infra", "system"}, false),
				Description:  "Specify the cluster profile type to use. Allowed values are `cluster`, `infra`, `add-on`, and `system`. These values map to the following User Interface (UI) labels. Use the value ' cluster ' for a **Full** cluster profile." + "For an Infrastructure cluster profile, use the value `infra`; for an Add-on cluster profile, use the value `add-on`." + "System cluster profiles can be specified using the value `system`. To learn more about cluster profiles, refer to the [Cluster Profile](https://docs.spectrocloud.com/cluster-profiles) documentation. Default value is `add-on`.",
				ForceNew:     true,
			},
			"profile_variables": schemas.ProfileVariables(),
			"pack":              schemas.PackSchema(),
		},
	}
}

// resourceClusterProfileCustomizeDiff is invoked at plan time. When the
// `immutable-clusterprofiles` feature_preview flag is enabled and the user is
// changing the `version` field on an existing resource, mark the field as
// `ForceNew` so Terraform plans a replacement (destroy + create) instead of an
// in-place update.
//
// Marking an attribute change as `ForceNew` from `CustomizeDiff` is the standard
// Terraform Plugin SDK v2 idiom for converting "this attribute changed" into
// "this resource must be replaced". We do it conditionally here because
// `spectrocloud_cluster_profile` has to preserve its legacy in-place mutation
// behavior for existing users who don't opt into the new flag -- gating on
// `CustomizeDiff` is the standard way to have one resource type with two
// different lifecycles in SDK v2.
//
// Combined with `lifecycle { create_before_destroy = true }` and `skip_destroy = true`
// in user HCL, the user gets one block per profile, a mutable `version` field
// from their HCL perspective, clean `git diff` between tags, and immutable
// preservation of old versions in Palette -- all while respecting Terraform Plugin
// SDK v2's contract that a resource's primary id is stable across in-place
// updates (which is why this approach has none of the stale-output or
// "[WARN] tolerating it because it is using the legacy plugin SDK" issues that
// come from trying to mutate `d.Id()` mid-Update).
//
// Validation happens in two shapes when the `immutable-clusterprofiles` flag is
// enabled:
//
//  1. Version bump (`version` field changing): mark `version` as `ForceNew` so
//     Terraform plans a replacement, AND require `skip_destroy = true` on the
//     resource. Without `skip_destroy`, the replacement's Delete phase would call
//     the Palette DELETE API and destroy the previous version -- defeating the
//     whole point of immutable versioning. Since `skip_destroy` is a
//     provider-schema attribute we can read at plan time, surfacing this as a
//     plan error rather than a runtime surprise is strictly better UX. The
//     companion `lifecycle { create_before_destroy = true }` block cannot be
//     validated from the provider (Terraform core parses lifecycle blocks before
//     the provider sees the diff), but the error message includes it so users
//     get both knobs from a single diagnostic.
//
//  2. Content change WITHOUT a version bump (any of `name`, `tags`, `pack`,
//     `description`, `profile_variables` changing while `version` stays the
//     same): reject with a plan-time error. The whole point of the flag is that
//     published cluster profile versions are immutable -- a user who edits the
//     pack content of v1.0.0 and re-applies without bumping the version is
//     asking the provider to silently mutate what's supposed to be an immutable
//     object. Without this check, the legacy `Update` path below would happily
//     send a PUT to the Palette API and the mutation would succeed silently
//     (the Palette API currently does not enforce version immutability
//     server-side). That would be the same class of bug as the
//     `clone-on-version-change` stale-output issue this whole PR was written to
//     fix -- a documented invariant that the code doesn't enforce.
func resourceClusterProfileCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if d.Id() == "" {
		// New resource -- no in-place update to convert
		return nil
	}
	if !isFeaturePreviewEnabled("immutable-clusterprofiles") {
		return nil
	}

	if d.HasChange("version") {
		// Version bump: mark ForceNew so Terraform plans a replacement, and
		// require the skip_destroy knob so the replacement's Delete phase
		// preserves the old version in Palette.
		if err := d.ForceNew("version"); err != nil {
			return err
		}
		if !d.Get("skip_destroy").(bool) {
			return fmt.Errorf(
				"immutable-clusterprofiles: version changes on %q require skip_destroy = true and "+
					"lifecycle { create_before_destroy = true } on the resource, so that the previous "+
					"version is preserved in Palette while Terraform replaces the resource. Add both to "+
					"your resource block:\n\n"+
					"  resource \"spectrocloud_cluster_profile\" %q {\n"+
					"    # ...\n"+
					"    skip_destroy = true\n\n"+
					"    lifecycle {\n"+
					"      create_before_destroy = true\n"+
					"    }\n"+
					"  }\n\n"+
					"This follows the standard Terraform Plugin SDK v2 immutable-versioned-resource "+
					"pattern used by aws_lambda_layer_version and similar resources. See the "+
					"\"Immutable versioning\" section of the spectrocloud_cluster_profile docs for a "+
					"full example.",
				d.Get("name").(string),
				d.Get("name").(string),
			)
		}
		return nil
	}

	// Version did NOT change, but the flag is on. Look for any content changes
	// that would have silently mutated the supposedly-immutable published
	// version. If any content field has changed, reject the plan.
	contentFields := []string{"name", "tags", "pack", "description", "profile_variables"}
	var changedContentFields []string
	for _, f := range contentFields {
		if d.HasChange(f) {
			changedContentFields = append(changedContentFields, f)
		}
	}
	if len(changedContentFields) > 0 {
		return fmt.Errorf(
			"immutable-clusterprofiles: cluster profile versions are immutable. "+
				"Detected changes to field(s) %v on %q (version %q) without a corresponding "+
				"version bump. To push these changes, increment the version field "+
				"(e.g., %q -> %q) so the provider creates a new immutable version via the "+
				"clone endpoint while the previous version is preserved in Palette. If you "+
				"intentionally want to mutate the existing version in place, remove the "+
				"\"immutable-clusterprofiles\" feature_preview flag from your provider block "+
				"(note: this reverts to the legacy destructive PUT behavior and is not "+
				"recommended).",
			changedContentFields,
			d.Get("name").(string),
			d.Get("version").(string),
			d.Get("version").(string),
			bumpPatchHint(d.Get("version").(string)),
		)
	}

	return nil
}

// bumpPatchHint returns a best-effort "next patch version" suggestion for the
// error message, purely cosmetic. If the version string isn't in a recognizable
// semver shape, it returns "<next-version>" as a placeholder. This is only used
// to make the error message more concrete -- the provider does not enforce any
// particular version format.
func bumpPatchHint(current string) string {
	// Cheap heuristic: if the last segment is an integer, bump it by 1.
	// Otherwise fall back to a placeholder.
	parts := strings.Split(current, ".")
	if len(parts) == 0 {
		return "<next-version>"
	}
	last := parts[len(parts)-1]
	n := 0
	if _, err := fmt.Sscanf(last, "%d", &n); err != nil || n < 0 {
		return "<next-version>"
	}
	parts[len(parts)-1] = fmt.Sprintf("%d", n+1)
	return strings.Join(parts, ".")
}

// findAnyExistingProfileVersionUID returns the UID of any existing version of
// the given profile name in the current scope, or an empty string if none exists.
//
// Used by the `immutable-clusterprofiles` Create path to find a clone source:
// when the user bumps the `version` field, Terraform plans a replacement
// (destroy + create), and the new resource's Create function runs against an
// existing Palette lineage. To produce the new immutable version, Create needs
// any existing uid in that lineage to call `CloneClusterProfile` against -- the
// exact source version doesn't matter, since clone always produces a new uid
// that we then overwrite with the user's pack content.
//
// This helper exists because Palette's data model treats every profile version
// as a separate object with its own UID -- there is no "lineage parent" object,
// so finding a clone source is a name-lookup across the listing endpoint rather
// than a parent-child traversal.
func findAnyExistingProfileVersionUID(c *client.V1Client, name string) (string, error) {
	profiles, err := c.GetClusterProfiles()
	if err != nil {
		return "", err
	}
	for _, p := range profiles {
		if p.Metadata != nil && p.Metadata.Name == name {
			return p.Metadata.UID, nil
		}
	}
	return "", nil
}

func resourceClusterProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// immutable-clusterprofiles: this is the Create half of the standard Terraform
	// Plugin SDK v2 replacement lifecycle for immutable-versioned resources. When
	// the user bumps `version` on an existing resource, CustomizeDiff marks the
	// field as ForceNew, Terraform plans destroy + create, and (with the user's
	// `lifecycle { create_before_destroy = true }`) the new resource's Create runs
	// FIRST, against an existing Palette lineage. We need to produce a new
	// immutable version of that lineage; the way Palette models this is via a
	// clone of any existing version object. So we look up an existing version uid
	// for the lineage by name, call CloneClusterProfile against it to get the new
	// version's uid, then apply the user's HCL content to the cloned object via
	// the same UpdateClusterProfile + PatchClusterProfile + PublishClusterProfile
	// chain that the in-place Update path uses for non-version field changes.
	//
	// The Terraform resource id is set once at Create time (via d.SetId) and never
	// mutates again, which is the SDK v2 contract -- and it's what makes outputs
	// against `.id` correct in the post-apply state without needing
	// `terraform apply -refresh-only`.
	if isFeaturePreviewEnabled("immutable-clusterprofiles") {
		name := d.Get("name").(string)
		version := d.Get("version").(string)

		// Check if the exact (name, version) already exists -- if so, adopt it.
		// This handles re-applies, multi-workspace patterns where two state files
		// declare the same profile, and the case where the user manually created
		// the version via the Palette UI before applying. Adopting an existing
		// uid into Terraform state is the standard SDK v2 pattern for handling
		// "create against an existing object".
		existingExactUID, lookupErr := c.GetClusterProfileUID(name, version)
		if lookupErr == nil && existingExactUID != "" {
			log.Printf("immutable-clusterprofiles: profile %s version %s already exists (UID %s), adopting (SDK v2 adopt-on-create pattern)", name, version, existingExactUID)
			d.SetId(existingExactUID)
			resourceClusterProfileRead(ctx, d, m)
			return diags
		}

		// Look for ANY existing version of this profile lineage to clone from.
		// We don't care which version we clone from -- clone always produces a new
		// uid that we then overwrite with the user's HCL content via the standard
		// update chain below.
		sourceUID, err := findAnyExistingProfileVersionUID(c, name)
		if err != nil {
			return diag.FromErr(err)
		}
		if sourceUID != "" {
			log.Printf("immutable-clusterprofiles: cloning profile %s to create version %s from existing lineage source UID %s (SDK v2 ForceNew replacement Create path)", name, version, sourceUID)
			cloneEntity := &models.V1ClusterProfileCloneEntity{
				Metadata: &models.V1ClusterProfileCloneMetaInputEntity{
					Name:    &name,
					Version: version,
				},
			}
			newUID, err := c.CloneClusterProfile(sourceUID, cloneEntity)
			if err != nil {
				return diag.FromErr(err)
			}
			d.SetId(newUID)

			// Apply the user's pack/tags/description to the cloned version. The
			// Palette clone API copies content from the source version, not from
			// the user's HCL -- we have to overwrite it with the desired content
			// using the standard Update + Patch + Publish chain.
			cp, err := c.GetClusterProfile(newUID)
			if err != nil {
				return diag.FromErr(err)
			}
			cluster, err := toClusterProfileUpdateWithResolution(d, cp, c)
			if err != nil {
				return diag.FromErr(err)
			}
			metadata, err := toClusterProfilePatch(d)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := c.UpdateClusterProfile(cluster); err != nil {
				return diag.FromErr(err)
			}
			if err := c.PatchClusterProfile(cluster, metadata); err != nil {
				return diag.FromErr(err)
			}
			if err := c.PublishClusterProfile(newUID); err != nil {
				return diag.FromErr(err)
			}

			resourceClusterProfileRead(ctx, d, m)
			return diags
		}
		// No prior version of this lineage exists -- fall through to the regular
		// Create path below to create the very first version. This is the
		// "first version of a brand new profile" case.
		log.Printf("immutable-clusterprofiles: no prior version of %s exists, falling through to fresh Create", name)
	}

	clusterProfile, err := toClusterProfileCreateWithResolution(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateClusterProfile(clusterProfile)
	adopted := false
	if err != nil {
		if !isFeaturePreviewEnabled("immutable-clusterprofiles") {
			return diag.FromErr(err)
		}
		// SDK v2 adopt-on-create pattern: if the profile already exists in Palette
		// (e.g. another root module created it, or it was created via the UI),
		// adopt the existing uid into Terraform state instead of failing. This
		// supports multi-environment patterns where multiple root modules declare
		// the same profile.
		name := d.Get("name").(string)
		version := d.Get("version").(string)
		existingUID, lookupErr := c.GetClusterProfileUID(name, version)
		if lookupErr != nil || existingUID == "" {
			return diag.FromErr(err) // return the original create error
		}
		log.Printf("immutable-clusterprofiles: profile %s version %s already exists (UID %s), adopting (SDK v2 adopt-on-create pattern)", name, version, existingUID)
		uid = existingUID
		adopted = true
	}

	// Publish only for newly created profiles -- adopted profiles are already published.
	if !adopted {
		if err = c.PublishClusterProfile(uid); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(uid)
	resourceClusterProfileRead(ctx, d, m)
	return diags
}

func resourceClusterProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	var diags diag.Diagnostics

	cp, err := c.GetClusterProfile(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if cp == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	err = flattenClusterProfileCommon(d, cp)
	if err != nil {
		diag.FromErr(err)
	}

	tags := flattenTags(cp.Metadata.Labels)
	if tags != nil {
		if err := d.Set("tags", tags); err != nil {
			return diag.FromErr(err)
		}
	}

	packManifests, d2, done2 := getPacksContent(cp.Spec.Published.Packs, c, d)
	if done2 {
		return d2
	}

	err = d.Set("name", cp.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	// Profile variables
	profileVariables, err := c.GetProfileVariables(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	pv, err := flattenProfileVariables(d, profileVariables)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("profile_variables", pv)
	if err != nil {
		return diag.FromErr(err)
	}

	diagPacks, diagnostics, done := GetDiagPacks(d, err)
	if done {
		return diagnostics
	}
	// Build registry maps to track which packs use registry_name or registry_uid
	registryNameMap := buildPackRegistryNameMap(d)
	registryUIDMap := buildPackRegistryUIDMap(d)
	packs, err := flattenPacksWithRegistryMaps(c, diagPacks, cp.Spec.Published.Packs, packManifests, registryNameMap, registryUIDMap)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pack", packs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenClusterProfileCommon(d *schema.ResourceData, cp *models.V1ClusterProfile) error {
	// set cloud
	if err := d.Set("cloud", cp.Spec.Published.CloudType); err != nil {
		return err
	}
	// set type
	if err := d.Set("type", cp.Spec.Published.Type); err != nil {
		return err
	}
	// set version
	if err := d.Set("version", cp.Spec.Published.ProfileVersion); err != nil {
		return err
	}
	return nil
}

func resourceClusterProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("profile_variables") {
		pvs, err := toClusterProfileVariables(d)
		if err != nil {
			return diag.FromErr(err)
		}
		mVars := &models.V1Variables{
			Variables: pvs,
		}
		err = c.UpdateProfileVariables(mVars, d.Id())
		if err != nil {
			oldVariables, _ := d.GetChange("profile_variables")
			_ = d.Set("profile_variables", oldVariables)
			return diag.FromErr(err)
		}
	}

	// Note on interaction with `immutable-clusterprofiles`:
	//
	// When the flag is enabled, CustomizeDiff catches both version changes
	// (by marking `version` as ForceNew, routing the work through
	// Create + skip_destroy-preserved Delete) AND content changes without
	// a version bump (by returning a plan-time error that tells the user
	// to bump the version or disable the flag). Either way, this Update
	// block is NOT reached when the flag is set -- replacement-based
	// work goes through Create, and disallowed mutations are rejected
	// at plan time before Update runs.
	//
	// This block is therefore the legacy in-place update path for users
	// who have NOT opted into the flag. It preserves the original
	// destructive PUT-based behavior (mutating the existing UID via
	// UpdateClusterProfile) for backward compatibility with existing CI
	// and HCL that was written against the pre-flag provider.
	if d.HasChanges("name") || d.HasChanges("tags") || d.HasChanges("pack") || d.HasChanges("description") || d.HasChanges("version") {
		log.Printf("Updating packs")
		cp, err := c.GetClusterProfile(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		cluster, err := toClusterProfileUpdateWithResolution(d, cp, c)
		if err != nil {
			return diag.FromErr(err)
		}
		metadata, err := toClusterProfilePatch(d)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := c.UpdateClusterProfile(cluster); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PatchClusterProfile(cluster, metadata); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PublishClusterProfile(cluster.Metadata.UID); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterProfileRead(ctx, d, m)

	return diags
}

func resourceClusterProfileDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	var diags diag.Diagnostics

	// skip_destroy: when set, removing the resource from Terraform state does NOT
	// call the Palette delete API. This is the standard Terraform Plugin SDK v2
	// preservation pattern for immutable-versioned resources. With `skip_destroy = true`,
	// version-bump replacements (triggered by ForceNew via the `immutable-clusterprofiles`
	// feature flag) and `lifecycle { create_before_destroy = true }` preserve old
	// versions in Palette while Terraform's state advances cleanly to the new version.
	// This is the canonical SDK v2 idiom for "I want my upstream system to keep
	// historical versions even though Terraform's state only tracks the latest one".
	if d.Get("skip_destroy").(bool) {
		log.Printf("skip_destroy: removing cluster profile %s from Terraform state without deleting from Palette (SDK v2 immutable-versioned-resource preservation pattern)", d.Id())
		return diags
	}

	if err := c.DeleteClusterProfile(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toClusterProfileCreateWithResolution(d *schema.ResourceData, c *client.V1Client) (*models.V1ClusterProfileEntity, error) {
	cp := toClusterProfileBasic(d)

	packs := make([]*models.V1PackManifestEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		if p, e := toClusterProfilePackCreateWithResolution(pack, c); e != nil {
			return nil, e
		} else {
			packs = append(packs, p)
		}
	}
	cp.Spec.Template.Packs = packs
	if profileVariable, err := toClusterProfileVariables(d); err == nil {
		cp.Spec.Variables = profileVariable
	} else {
		return cp, err
	}
	return cp, nil
}

func toClusterProfileBasic(d *schema.ResourceData) *models.V1ClusterProfileEntity {
	description := ""
	if d.Get("description") != nil {
		description = d.Get("description").(string)
	}
	cp := &models.V1ClusterProfileEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
			Annotations: map[string]string{
				"description": description,
			},
			Labels: toTags(d),
		},
		Spec: &models.V1ClusterProfileEntitySpec{
			Template: &models.V1ClusterProfileTemplateDraft{
				CloudType: d.Get("cloud").(string),
				Type:      types.Ptr(models.V1ProfileType(d.Get("type").(string))),
			},
			Version: d.Get("version").(string),
		},
	}
	return cp
}

func toClusterProfilePackCreate(pSrc interface{}) (*models.V1PackManifestEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pUID := p["uid"].(string)
	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pRegistryName := ""
	if p["registry_name"] != nil {
		pRegistryName = p["registry_name"].(string)
	}
	pType := models.V1PackType(p["type"].(string))

	// Validate pack UID or resolution fields
	if err := schemas.ValidatePackUIDOrResolutionFields(p); err != nil {
		return nil, err
	}

	// Note: registry_name is stored but not resolved here since we don't have client
	// Actual resolution will happen in toClusterProfilePackCreateWithResolution

	switch pType {
	case models.V1PackTypeSpectro:
		if pUID == "" {
			// UID not provided, validation already passed, so we have all resolution fields
			// This path should not be reached if validation is working correctly
			if pTag == "" || (pRegistryUID == "" && pRegistryName == "") {
				return nil, fmt.Errorf("pack %s: internal error - validation should have caught missing resolution fields", pName)
			}
		}
	case models.V1PackTypeManifest:
		if pUID == "" {
			pUID = "spectro-manifest-pack"
		}
	}

	pack := &models.V1PackManifestEntity{
		Name:        types.Ptr(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        &pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}

	manifests := make([]*models.V1ManifestInputEntity, 0)
	if len(p["manifest"].([]interface{})) > 0 {
		for _, manifest := range p["manifest"].([]interface{}) {
			m := manifest.(map[string]interface{})
			manifests = append(manifests, &models.V1ManifestInputEntity{
				Content: strings.TrimSpace(m["content"].(string)),
				Name:    m["name"].(string),
			})
		}
	}
	pack.Manifests = manifests

	return pack, nil
}

func toClusterProfilePackCreateWithResolution(pSrc interface{}, c *client.V1Client) (*models.V1PackManifestEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pUID := p["uid"].(string)
	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pRegistryName := ""
	if p["registry_name"] != nil {
		pRegistryName = p["registry_name"].(string)
	}
	pType := models.V1PackType(p["type"].(string))

	// Validate pack UID or resolution fields
	if err := schemas.ValidatePackUIDOrResolutionFields(p); err != nil {
		return nil, err
	}

	// If registry_name is provided, resolve it to registry_uid
	if pRegistryName != "" && pRegistryUID == "" {
		resolvedUID, err := resolveRegistryNameToUID(c, pRegistryName, p["type"].(string))
		if err != nil {
			return nil, fmt.Errorf("pack %s: %w", pName, err)
		}
		pRegistryUID = resolvedUID
	}

	switch pType {
	case models.V1PackTypeOci, models.V1PackTypeSpectro, models.V1PackTypeHelm:
		if pUID == "" {
			isSyncSupported, found := getRegistryIsSyncSupported(c, pRegistryUID, pType)
			if found && isSyncSupported {
				resolvedUID, err := resolvePackUID(c, pName, pTag, pRegistryUID)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve pack UID for pack %s: %w", pName, err)
				}
				pUID = resolvedUID
			}
			// Else skip pack UID resolution when sync is not supported or registry not found
		}
	case models.V1PackTypeManifest:
		if pUID == "" {
			pUID = "spectro-manifest-pack"
		}
	}

	pack := &models.V1PackManifestEntity{
		Name:        types.Ptr(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        &pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}

	manifests := make([]*models.V1ManifestInputEntity, 0)
	if len(p["manifest"].([]interface{})) > 0 {
		for _, manifest := range p["manifest"].([]interface{}) {
			m := manifest.(map[string]interface{})
			manifests = append(manifests, &models.V1ManifestInputEntity{
				Content: strings.TrimSpace(m["content"].(string)),
				Name:    m["name"].(string),
			})
		}
	}
	pack.Manifests = manifests

	return pack, nil
}

func toClusterProfileUpdateWithResolution(d *schema.ResourceData, cluster *models.V1ClusterProfile, c *client.V1Client) (*models.V1ClusterProfileUpdateEntity, error) {
	cp := &models.V1ClusterProfileUpdateEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1ClusterProfileUpdateEntitySpec{
			Template: &models.V1ClusterProfileTemplateUpdate{
				Type: types.Ptr(models.V1ProfileType(d.Get("type").(string))),
			},
			Version: d.Get("version").(string),
		},
	}
	packs := make([]*models.V1PackManifestUpdateEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		if p, e := toClusterProfilePackUpdateWithResolution(pack, cluster.Spec.Published.Packs, c); e != nil {
			return nil, e
		} else {
			packs = append(packs, p)
		}
	}
	cp.Spec.Template.Packs = packs

	return cp, nil
}

func toClusterProfilePatch(d *schema.ResourceData) (*models.V1ProfileMetaEntity, error) {
	description := ""
	if d.Get("description") != nil {
		description = d.Get("description").(string)
	}
	metadata := &models.V1ProfileMetaEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name: d.Get("name").(string),
			Annotations: map[string]string{
				"description": description,
			},
			Labels: toTags(d),
		},
		Spec: &models.V1ClusterProfileSpecEntity{
			Version: d.Get("version").(string),
		},
	}

	return metadata, nil
}

func toClusterProfilePackUpdateWithResolution(pSrc interface{}, packs []*models.V1PackRef, c *client.V1Client) (*models.V1PackManifestUpdateEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pUID := p["uid"].(string)

	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pType := models.V1PackType(p["type"].(string))

	// Validate pack UID or resolution fields
	if err := schemas.ValidatePackUIDOrResolutionFields(p); err != nil {
		return nil, err
	}

	switch pType {
	case models.V1PackTypeOci, models.V1PackTypeSpectro, models.V1PackTypeHelm:
		if pUID == "" {
			isSyncSupported, found := getRegistryIsSyncSupported(c, pRegistryUID, pType)
			if found && isSyncSupported {
				resolvedUID, err := resolvePackUID(c, pName, pTag, pRegistryUID)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve pack UID for pack %s: %w", pName, err)
				}
				pUID = resolvedUID
			}
			// Else skip pack UID resolution when sync is not supported or registry not found
		}
	case models.V1PackTypeManifest:
		if pUID == "" {
			pUID = "spectro-manifest-pack"
		}
	}

	pack := &models.V1PackManifestUpdateEntity{
		//Layer:  p["layer"].(string),
		Name:        types.Ptr(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        &pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}

	manifests := make([]*models.V1ManifestRefUpdateEntity, 0)
	for _, manifest := range p["manifest"].([]interface{}) {
		m := manifest.(map[string]interface{})
		manifests = append(manifests, &models.V1ManifestRefUpdateEntity{
			Content: strings.TrimSpace(m["content"].(string)),
			Name:    types.Ptr(m["name"].(string)),
			UID:     getManifestUID(m["name"].(string), packs),
		})
	}
	pack.Manifests = manifests

	return pack, nil
}

func getManifestUID(name string, packs []*models.V1PackRef) string {
	for _, pack := range packs {
		for _, manifest := range pack.Manifests {
			if manifest.Name == name {
				return manifest.UID
			}
		}
	}

	return ""
}

func toClusterProfileVariables(d *schema.ResourceData) ([]*models.V1Variable, error) {
	var profileVariables []*models.V1Variable
	if pVariables, ok := d.GetOk("profile_variables"); ok {
		if pVariables.([]interface{})[0] != nil {
			variables := pVariables.([]interface{})[0].(map[string]interface{})["variable"]
			for _, v := range variables.([]interface{}) {
				variable := v.(map[string]interface{})
				pv := &models.V1Variable{
					DefaultValue: getVariableString(variable, "default_value"),
					Description:  getVariableString(variable, "description"),
					DisplayName:  variable["display_name"].(string),
					Format:       types.Ptr(models.V1VariableFormat(getVariableString(variable, "format", "string"))),
					Hidden:       getVariableBool(variable, "hidden"),
					Immutable:    getVariableBool(variable, "immutable"),
					Name:         StringPtr(variable["name"].(string)),
					Regex:        getVariableString(variable, "regex"),
					IsSensitive:  getVariableBool(variable, "is_sensitive"),
					Required:     getVariableBool(variable, "required"),
					InputType:    types.Ptr(models.V1VariableInputType(getVariableString(variable, "input_type", "text"))),
					Options:      toVariableOptions(variable["options"]),
				}
				profileVariables = append(profileVariables, pv)
			}
		}
	}
	return profileVariables, nil
}

// getVariableString returns the string value for key from the variable map, or the first default if missing.
func getVariableString(variable map[string]interface{}, key string, defaultVal ...string) string {
	if v, ok := variable[key].(string); ok {
		return v
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// getVariableBool returns the bool value for key from the variable map, or false if missing.
func getVariableBool(variable map[string]interface{}, key string) bool {
	if v, ok := variable[key].(bool); ok {
		return v
	}
	return false
}

// toVariableOptions converts the Terraform options list to API V1VariableOption slice.
func toVariableOptions(options interface{}) []*models.V1VariableOption {
	if options == nil {
		return nil
	}
	list, ok := options.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}
	out := make([]*models.V1VariableOption, 0, len(list))
	for _, o := range list {
		opt, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		val, _ := opt["value"].(string)
		label, _ := opt["label"].(string)
		desc, _ := opt["description"].(string)
		// default is computed in schema; use when present (e.g. from state after read)
		defaultOpt := false
		if d, ok := opt["default"].(bool); ok {
			defaultOpt = d
		}
		out = append(out, &models.V1VariableOption{
			Value:       types.Ptr(val),
			Label:       label,
			Description: desc,
			Default:     defaultOpt,
		})
	}
	return out
}

// flattenVariableOptions converts API V1VariableOption slice to Terraform options list.
func flattenVariableOptions(opts []*models.V1VariableOption) []interface{} {
	if len(opts) == 0 {
		return nil
	}
	out := make([]interface{}, 0, len(opts))
	for _, o := range opts {
		if o == nil {
			continue
		}
		m := map[string]interface{}{
			"default":     o.Default,
			"description": o.Description,
			"label":       o.Label,
			"value":       String(o.Value),
		}
		out = append(out, m)
	}
	return out
}

func flattenProfileVariables(d *schema.ResourceData, pv []*models.V1Variable) ([]interface{}, error) {
	if len(pv) == 0 {
		return make([]interface{}, 0), nil
	}
	var variables []interface{}
	for _, v := range pv {
		variable := make(map[string]interface{})
		variable["name"] = v.Name
		variable["display_name"] = v.DisplayName
		variable["description"] = v.Description
		variable["format"] = v.Format
		variable["default_value"] = v.DefaultValue
		variable["regex"] = v.Regex
		variable["required"] = v.Required
		variable["immutable"] = v.Immutable
		variable["hidden"] = v.Hidden
		variable["is_sensitive"] = v.IsSensitive
		if v.InputType != nil {
			variable["input_type"] = string(*v.InputType)
		} else {
			variable["input_type"] = "text"
		}
		variable["options"] = flattenVariableOptions(v.Options)
		variables = append(variables, variable)
	}
	// Sorting ordering the list per configuration this reference if we need to change profile_variables to TypeList
	var sortedVariables []interface{}
	var configVariables []interface{}
	if v, ok := d.GetOk("profile_variables"); ok {
		configVariables = v.([]interface{})[0].(map[string]interface{})["variable"].([]interface{})
		for _, cv := range configVariables {
			mapV := cv.(map[string]interface{})
			for _, va := range variables {
				vs := va.(map[string]interface{})
				if mapV["name"].(string) == String(vs["name"].(*string)) {
					sortedVariables = append(sortedVariables, va)
				}
			}
		}
	} else {
		sortedVariables = variables
	}

	flattenProVariables := make([]interface{}, 1)
	flattenProVariables[0] = map[string]interface{}{
		"variable": sortedVariables,
	}
	return flattenProVariables, nil
}

// getRegistryIsSyncSupported returns (isSyncSupported, found). Resolve pack UID only when found and isSyncSupported is true.
func getRegistryIsSyncSupported(c *client.V1Client, registryUID string, packType models.V1PackType) (isSyncSupported bool, found bool) {
	if registryUID == "" {
		return false, false
	}
	var status *models.V1RegistrySyncStatus
	var err error
	switch packType {
	case models.V1PackTypeHelm:
		status, err = c.GetHelmRegistrySyncStatus(registryUID)
	case models.V1PackTypeOci:
		status, err = c.GetOciBasicRegistrySyncStatus(registryUID)
	case models.V1PackTypeSpectro:
		status, err = c.GetPackRegistrySyncStatus(registryUID)
	default:
		return false, false
	}
	if err != nil || status == nil {
		return false, false
	}
	return status.IsSyncSupported, true
}

// resolvePackUID resolves the pack UID based on name, tag, and registry_uid
func resolvePackUID(c *client.V1Client, name, tag, registryUID string) (string, error) {
	if name == "" || tag == "" || registryUID == "" {
		return "", fmt.Errorf("name, tag, and registry_uid are all required for pack resolution")
	}

	// Get pack versions by name and registry
	packVersions, err := c.GetPacksByNameAndRegistry(name, registryUID)
	if err != nil {
		return "", fmt.Errorf("failed to get pack versions for name %s in registry %s: %w", name, registryUID, err)
	}

	if packVersions == nil || len(packVersions.Tags) == 0 {
		return "", fmt.Errorf("no pack found with name %s in registry %s", name, registryUID)
	}

	// Find the pack with matching tag/version
	for _, packTag := range packVersions.Tags {
		if packTag.Version == tag {
			return packTag.PackUID, nil
		}
	}

	return "", fmt.Errorf("no pack found with name %s, tag %s in registry %s", name, tag, registryUID)
}
