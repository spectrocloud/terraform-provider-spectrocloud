# CLAUDE.md — terraform-provider-spectrocloud

## What This Repo Is

This is a Terraform provider (terraform-plugin-sdk v2) for the Spectro Cloud Palette platform. It manages Kubernetes clusters across multiple cloud providers (AWS, Azure, GCP, vSphere, OpenStack, MAAS, EKS, AKS, GKE, Edge, etc.), cluster profiles, cloud accounts, projects, users, workspaces, and platform configuration. All API calls go through `github.com/spectrocloud/palette-sdk-go/client`, referred to internally as "hapi" or the V1 client.

---

## File Naming Conventions

- Resources: `spectrocloud/resource_<thing>.go` (e.g., `resource_cluster_aws.go`, `resource_project.go`)
- Import handlers: `spectrocloud/resource_<thing>_import.go` — always a separate file
- Data sources: `spectrocloud/data_source_<thing>.go`
- Shared cluster logic: `spectrocloud/cluster_common_<aspect>.go` (e.g., `cluster_common_crud.go`, `cluster_common_fields.go`, `cluster_common_profiles.go`, `cluster_common_hash.go`)
- Reusable schema blocks: `spectrocloud/schemas/<block>.go` (e.g., `schemas/pack.go`, `schemas/cluster_profile.go`, `schemas/backup_policy.go`)
- Provider registration: `spectrocloud/provider.go`
- Helper utils: `spectrocloud/utils.go`, `spectrocloud/common_utils.go`
- Tests live alongside their source file: `resource_foo_test.go` or `data_source_foo_test.go`
- Integration-style tests: `tests/` directory, organized by feature

---

## Resource Definition Pattern

Every resource follows this exact skeleton:

```go
func resourceFoo() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceFooCreate,
        ReadContext:   resourceFooRead,
        UpdateContext: resourceFooUpdate,
        DeleteContext: resourceFooDelete,
        Importer: &schema.ResourceImporter{
            StateContext: resourceFooImport,
        },
        Description: "One-sentence description of what this manages.",

        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(10 * time.Minute),
            Update: schema.DefaultTimeout(10 * time.Minute),
            Delete: schema.DefaultTimeout(10 * time.Minute),
        },

        SchemaVersion: 2,
        Schema: map[string]*schema.Schema{
            // fields here
        },
    }
}
```

Clusters use 60-minute timeouts. Non-cluster resources use 10-20 minutes. `SchemaVersion` is typically set to 2 on resources that have been through schema migrations. Not all resources have it; only add it when the resource already has it or when a migration is being done.

Register the new resource in `provider.go` in `ResourcesMap` under a `"spectrocloud_<name>"` key.

---

## Schema Patterns

**Standard fields that nearly every resource has:**

```go
"name": {
    Type:     schema.TypeString,
    Required: true,
},
"context": {
    Type:         schema.TypeString,
    Optional:     true,
    Default:      "project",
    ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
    Description: "The context of the resource. Allowed values are `project` or `tenant`. " +
        "Default value is `project`. " + PROJECT_NAME_NUANCE,
},
"tags": {
    Type:     schema.TypeSet,
    Optional: true,
    Set:      schema.HashString,
    Elem: &schema.Schema{
        Type: schema.TypeString,
    },
    Description: "A list of tags to be applied to the resource. Tags must be in the form of `key:value`.",
},
"description": {
    Type:     schema.TypeString,
    Optional: true,
    Default:  "",
},
```

**Computed-only fields (server-assigned IDs, kubeconfig, etc.):**

```go
"cloud_config_id": {
    Type:        schema.TypeString,
    Computed:    true,
    Description: "...",
    Deprecated:  "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
},
```

**Embedded object blocks use TypeList with MaxItems: 1:**

```go
"cloud_config": {
    Type:     schema.TypeList,
    ForceNew: true,
    Required: true,
    MaxItems: 1,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            // fields
        },
    },
},
```

**Sets of objects (machine pools, etc.) use TypeSet with a hash function:**

```go
"machine_pool": {
    Type:     schema.TypeSet,
    Required: true,
    Set:      resourceMachinePoolAwsHash,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            // fields
        },
    },
},
```

**Reusable schema blocks are in `spectrocloud/schemas/` and called as functions:**

```go
"cluster_profile": schemas.ClusterProfileSchema(),
"backup_policy":   schemas.BackupPolicySchema(),
"scan_policy":     schemas.ScanPolicySchema(),
"taints":          schemas.ClusterTaintsSchema(),
"node":            schemas.NodeSchema(),
```

**DiffSuppressFunc for YAML/multi-line strings:**

```go
DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
    return strings.TrimSpace(old) == strings.TrimSpace(new)
},
```

**ValidateFunc for enums:**

```go
ValidateFunc: validation.StringInSlice([]string{"on-demand", "spot"}, false),
```

Every schema field must have a `Description`. No exceptions — undocumented fields will cause doc generation to fail.

---

## CRUD Implementation Patterns

**Create:**

```go
func resourceFooCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    resourceContext := d.Get("context").(string)
    c := getV1ClientWithResourceContext(m, resourceContext)

    var diags diag.Diagnostics

    foo, err := toFoo(d)
    if err != nil {
        return diag.FromErr(err)
    }

    uid, err := c.CreateFoo(foo)
    if err != nil {
        return diag.FromErr(err)
    }

    d.SetId(uid)

    resourceFooRead(ctx, d, m)

    return diags
}
```

Key rules:
- Always call `getV1ClientWithResourceContext(m, resourceContext)` — never cast `m` directly except inside `getV1ClientWithResourceContext`.
- `var diags diag.Diagnostics` is declared but typically returned empty at the end; real errors come via `diag.FromErr(err)`.
- After create, always call the Read function to sync state. Do not return the diags from the Read call — the pattern is just `resourceFooRead(ctx, d, m)` and then `return diags`.
- Clusters need `waitForClusterCreation(ctx, d, uid, diags, c, true)` before the Read call.

**Read:**

```go
//goland:noinspection GoUnhandledErrorResult
func resourceFooRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    resourceContext := d.Get("context").(string)
    c := getV1ClientWithResourceContext(m, resourceContext)

    var diags diag.Diagnostics

    foo, err := c.GetFoo(d.Id())
    if err != nil {
        return handleReadError(d, err, diags)
    } else if foo == nil {
        // Deleted - Terraform will recreate it
        d.SetId("")
        return diags
    }

    if err := d.Set("name", foo.Metadata.Name); err != nil {
        return diag.FromErr(err)
    }
    // ... set other fields

    return diags
}
```

Key rules:
- The `_ context.Context` parameter (context ignored) is standard for Read when no async waiting is needed. Use named `ctx` only when actually passing it somewhere.
- The `// Deleted - Terraform will recreate it` comment is canonical — include it on the `d.SetId("")` line.
- Use `handleReadError(d, err, diags)` — it handles 404s by clearing the ID vs. returning a real error.
- Every `d.Set(...)` call result must be checked: `if err := d.Set(...); err != nil { return diag.FromErr(err) }`.

**Update:**

```go
func resourceFooUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    resourceContext := d.Get("context").(string)
    c := getV1ClientWithResourceContext(m, resourceContext)

    var diags diag.Diagnostics

    if d.HasChanges("name", "description", "tags") {
        err := c.UpdateFoo(d.Id(), toFoo(d))
        if err != nil {
            return diag.FromErr(err)
        }
    }

    resourceFooRead(ctx, d, m)
    return diags
}
```

For machine pool changes, use the `d.GetChange("machine_pool")` pattern with old/new sets, building `osMap` and `nsMap` keyed by name, then create/update/delete pools individually. See `resource_cluster_aws.go` for the canonical example.

**Delete:**

```go
func resourceFooDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    resourceContext := d.Get("context").(string)
    c := getV1ClientWithResourceContext(m, resourceContext)

    var diags diag.Diagnostics

    err := c.DeleteFoo(d.Id())
    if err != nil {
        return diag.FromErr(err)
    }

    // d.SetId("") is automatically called assuming delete returns no errors

    return diags
}
```

The comment `// d.SetId("") is automatically called assuming delete returns no errors` is part of the style — include it.

---

## Error Handling Rules

- **All errors from the API go through `diag.FromErr(err)`** — never construct a `diag.Diagnostic` slice manually unless writing a custom summary/detail message.
- **Read errors use `handleReadError(d, err, diags)`** — this function checks `herr.IsNotFound(err)` and clears the ID for 404s rather than failing.
- **Manual diagnostics** (used sparingly, for validation failures or warnings) follow this form:
  ```go
  diags = append(diags, diag.Diagnostic{
      Severity: diag.Error,
      Summary:  "Force delete validation failed",
      Detail:   "`force_delete_delay` should not be greater than default delete timeout.",
  })
  return diags
  ```
- **Warnings** use `diag.Warning` severity. The `generalWarningForRepave(&diags)` pattern appends a standard warning about node repaves in cloud config flatten functions.
- **Nil checks before API calls**: always check that the response is not nil before accessing fields. Pattern: `if foo == nil { d.SetId(""); return diags }`.
- **Functions that return `(diag.Diagnostics, bool)`**: the bool is `true` when the caller should stop and return the diagnostics immediately (i.e., an error occurred). Check: `if done { return diagnostics }`.

---

## API Client Patterns

**Getting the client:**

```go
c := getV1ClientWithResourceContext(m, resourceContext)
```

This is always the first thing in every CRUD function. The `resourceContext` comes from `d.Get("context").(string)`. Resources that are always tenant-scoped (like users) pass `"tenant"` directly. Resources that are provider-level pass `""`.

**The client type is `*client.V1Client`** from `github.com/spectrocloud/palette-sdk-go/client`.

**Pointer helpers:**

- `types.Ptr(value)` — generic pointer wrapper from the local `types` package. Use this over `&localVar` when building model structs inline.
- `SafeInt32(intVal)` — converts int to int32 with overflow protection. Always use this when assigning to model fields typed `*int32` or `int32`.
- `SafeInt64(intVal)` — same for int64.
- `StringPtr(s)`, `BoolPtr(b)`, `Int32Ptr(i)` etc. — in `utils.go`, use these for simple pointer creation.

**Building API model structs:**

Follow the `to<Thing>` naming convention for functions that convert `*schema.ResourceData` into API model structs:

```go
func toFoo(d *schema.ResourceData) *models.V1FooEntity {
    return &models.V1FooEntity{
        Metadata: &models.V1ObjectMeta{
            Name:   d.Get("name").(string),
            UID:    d.Id(),
            Labels: toTags(d),
        },
        Spec: &models.V1FooSpec{
            // ...
        },
    }
}
```

For cluster resources, use `getClusterMetadata(d)` to build the `*models.V1ObjectMeta` block.

**Flattening API responses into state:**

Follow the `flatten<Thing>` naming convention:

```go
func flattenFoo(d *schema.ResourceData, foo *models.V1Foo) (diag.Diagnostics, bool) {
    if err := d.Set("name", foo.Metadata.Name); err != nil {
        return diag.FromErr(err), true
    }
    // ...
    return nil, false
}
```

Some flatten functions return `diag.Diagnostics` directly (no bool). Others return `(diag.Diagnostics, bool)` where the bool signals "done/error". Be consistent with the resource's existing pattern.

---

## State Refresh Patterns

For long-running operations (cluster create/delete), use `retry.StateChangeConf`:

```go
stateConf := &retry.StateChangeConf{
    Pending:    []string{"Pending", "Provisioning"},
    Target:     []string{"Running-Healthy"},
    Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
    Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
    MinTimeout: 10 * time.Second,
    Delay:      30 * time.Second,
}

_, err := stateConf.WaitForStateContext(ctx)
if err != nil {
    return diag.FromErr(err), true
}
```

The refresh function returns `(interface{}, string, error)`. Return `nil, "Deleted", nil` when the resource is gone. Return `nil, "", err` on API error. All state refresh functions live in `cluster_common_crud.go` for clusters.

Use the shared `waitForClusterCreation`, `waitForClusterDeletion`, `waitForClusterReady` functions — do not duplicate this logic.

---

## Import Support

Every resource has a matching `resource_<name>_import.go` file. The pattern is always:

```go
func resourceFooImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
    c, err := GetCommonFoo(d, m)
    if err != nil {
        return nil, err
    }

    diags := resourceFooRead(ctx, d, m)
    if diags.HasError() {
        return nil, fmt.Errorf("could not read foo for import: %v", diags)
    }

    return []*schema.ResourceData{d}, nil
}
```

For cluster imports, also call `flattenCommonAttributeForClusterImport(c, d)` to populate default Terraform-managed values that don't come back from the API.

The import ID format for scoped resources is `<context>:<uid>`. Parsing is done with `ParseResourceID(d)`, which sets `d.SetId(uid)` and returns `resourceContext` and `clusterID`.

---

## Data Source Patterns

Data sources only have `ReadContext`. No `CreateContext`, `UpdateContext`, or `DeleteContext`.

```go
func dataSourceFoo() *schema.Resource {
    return &schema.Resource{
        ReadContext: dataSourceFooRead,
        Description: "...",
        Schema: map[string]*schema.Schema{
            "id": {
                Type:         schema.TypeString,
                Optional:     true,
                Computed:     true,
                ExactlyOneOf: []string{"id", "name"},
                Description:  "...",
            },
            "name": {
                Type:         schema.TypeString,
                Optional:     true,
                Computed:     true,
                ExactlyOneOf: []string{"id", "name"},
                Description:  "...",
            },
            // output-only fields are Computed: true
        },
    }
}
```

Data sources typically look up by `id` or `name` using `ExactlyOneOf`. Output fields are `Computed: true`. Input/filter fields may be `Optional: true`. Data sources do not have `Importer` blocks or `Timeouts`.

Register in `provider.go` under `DataSourcesMap`.

---

## Tags Handling

Tags are stored as Terraform `TypeSet` of strings in the form `"key:value"` or just `"key"` (which maps to the value `"spectro__tag"`).

- `toTags(d)` converts the set to `map[string]string` for the API.
- `flattenTags(labels)` converts the API `map[string]string` back to `[]interface{}`.
- `tags_map` is the newer `TypeMap` alternative, mutually exclusive with `tags` via `ConflictsWith`.
- `toTagsMap(d)` and `flattenTagsMap(labels)` are used for `tags_map`.

---

## Hash Functions for TypeSet

Every `TypeSet` of objects needs a custom hash function. Naming convention: `resourceMachinePool<Cloud>Hash`. They all live in `cluster_common_hash.go`. Hash functions use `bytes.Buffer`, `fmt.Sprintf` to write fields, and `hash(buf.String())` (FNV-32a) to produce the int.

```go
func resourceMachinePoolFooHash(v interface{}) int {
    m := v.(map[string]interface{})
    buf := CommonHash(m) // handles standard fields: labels, taints, control_plane, name, count, etc.

    buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))

    return int(hash(buf.String()))
}
```

Use `CommonHash(m)` as the base for machine pools — it handles the shared fields. Add cloud-specific fields after.

---

## Naming Conventions

- Functions: `camelCase` throughout. No underscores except in test helpers.
- Context variables extracted from schema: name them `resourceContext` or `ProfileContext` or `clusterContext` — not standardized, match the existing file.
- Local variables for API responses: short descriptive names — `cluster`, `account`, `profile`, `foo`. Not `resp` or `result`.
- Map iteration for sets: `mp.(map[string]interface{})` → local variable `m`.
- Old/new change detection: `oraw, nraw := d.GetChange("machine_pool")` → `os := oraw.(*schema.Set)` → `osMap`, `nsMap`.
- Model types follow the palette-sdk-go convention: `models.V1<Something>Entity`, `models.V1<Something>Spec`, etc.

---

## Code Style Rules

- **No unused variables.** Go compiler will catch this, but `_` assignments are used when you need to discard a return value intentionally.
- **Type assertions are explicit**: always `thing.(string)`, `thing.(int)`, `thing.(bool)`, `thing.(map[string]interface{})`, `thing.(*schema.Set)`. Never use `fmt.Sprintf("%v", thing)` to coerce types.
- **Nil checks before accessing nested fields**: `if config.Spec != nil && config.Spec.ClusterConfig != nil { ... }`.
- **Comments on commented-out code**: leaving `//` blocks is acceptable (see the commented-out update pending states in `cluster_common_crud.go`). Don't clean up commented code that was intentionally preserved.
- **Inline comments** for non-obvious behavior: `// gnarly, I know! =/`, `// since known issue in TF SDK: ...`. These are genuine and appropriate.
- **`//goland:noinspection GoUnhandledErrorResult`** before Read functions that call other functions without checking their return — this is the IDE suppression comment, keep it.
- **Import grouping**: standard library first, then external packages, then internal packages. Use a blank line between groups.
- **No `gofmt` deviations** — the code is gofmt-clean. Keep it that way.
- **`sort.SliceStable`** is used for deterministic ordering of machine pools (control planes first, then alphabetical by name). Use `SliceStable`, not `SliceSort`.

---

## Testing Conventions

Tests use the `testify` library (`assert`, `require`). The test suite spins up a mock API server before running:

- `TestMain` in `common_test.go` starts/stops the mock server via shell scripts in `tests/mockApiServer/`.
- `unitTestMockAPIClient` and `unitTestMockAPINegativeClient` are package-level variables used by all tests.
- Unit tests use the table-driven pattern:

```go
func TestToFoo(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected *models.V1Foo
        wantErr  bool
    }{
        {
            name:  "ValidInput",
            input: map[string]interface{}{...},
            expected: &models.V1Foo{...},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := toFoo(tt.input)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

- Test functions that test CRUD operations directly call the `resource*Create/Read/Update/Delete` functions with `unitTestMockAPIClient` as the `m` argument.
- `assertFirstDiagMessage(t, diags, "expected message")` is a helper in `common_test.go` for checking diagnostic output.
- Acceptance tests for integration scenarios live in `tests/` and are meant to run against a real or mocked API server, not as part of `go test ./...` in CI without configuration.

---

## Patterns to Avoid

- **Do not** use `d.Set(...)` without checking the error. Always wrap in `if err := d.Set(...); err != nil { return diag.FromErr(err) }`.
- **Do not** cast `m` directly to `*client.V1Client` in CRUD functions. Always go through `getV1ClientWithResourceContext`.
- **Do not** create inline retry loops for API polling. Use `retry.StateChangeConf` from `github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry`.
- **Do not** write a `flattenX` or `toX` function that panics on nil. Guard with nil checks.
- **Do not** add `Computed: true` to fields that the user must supply. `Computed` means the API can set or change the value.
- **Do not** omit `Description` from any schema field.
- **Do not** use `schema.TypeList` for collections where ordering doesn't matter — use `schema.TypeSet` with a hash function instead.
- **Do not** duplicate the state-refresh polling logic — use the shared `waitForClusterCreation`, `waitForClusterDeletion` functions.
- **Do not** put import logic directly in the main resource file — it goes in `resource_<name>_import.go`.
- **Do not** write a data source with create/update/delete stubs. Data sources only have `ReadContext`.
- **Do not** use `fmt.Errorf` to wrap API errors in CRUD functions — use `diag.FromErr(err)` directly.
- **Do not** use generic variable names like `result`, `response`, `res` for API return values. Use the domain term: `cluster`, `profile`, `account`.

## Function Design & Testability

- **Every function does one thing and fits in ~20–30 lines.** If it grows beyond that, extract named helpers.
- **Write functions so they can be unit tested in isolation** — no hidden side effects, no global state access, no I/O buried inside business logic.
- **Most business logic must be unit testable** without spinning up a server, database, or Kubernetes cluster. Separate I/O at the boundary.
- **Use guard clauses / early returns** to reduce nesting. Flat code is easier to read and test than deeply nested.
- **Accept interfaces, return concrete types.** This makes callers mockable without reflection or code generation.
- **Keep interfaces small** — 1–3 methods. Large interfaces are hard to mock and signal poor separation of concerns.

## General Go Practices

- **Dependency injection over globals.** Pass dependencies via constructors or function parameters — not package-level singletons (except logging).
- **`context.Context` is always the first parameter** on any function that performs I/O. Never store it in a struct field.
- **Table-driven tests** for any function with multiple input/output cases: `[]struct{ name, input, expected }` with `t.Run`.
- **Test naming:** `TestFuncName_Scenario` — e.g. `TestCreateCluster_MissingName`.
- **Prefer `switch` over long `if/else if` chains.**
- **Short variable names in small scopes** (`i`, `v`, `err`) are idiomatic; use descriptive names in wider scopes.
- **No goroutines unless concurrency is genuinely required.** Sequential code is easier to test and reason about.
- **Avoid `init()` for anything except registering handlers or loggers.** Never use it for config loading or side-effectful setup.
- **Respect context cancellation** in any loop that calls external services.
- **Import grouping:** stdlib / external / internal — separated by blank lines, sorted by `goimports`.
- **Don't over-abstract.** Don't create an interface or wrapper until there are ≥2 concrete implementations or a clear testing need.
- **No naked `panic` in library code.** Panics are only acceptable in `main` or test setup for truly unrecoverable state.
