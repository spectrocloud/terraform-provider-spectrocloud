# Terraform Provider Resource Patterns

## Common Resource Patterns

### Standard Resource Template

```go
package spectrocloud

import (
    "context"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceExample struct {
    provider *Provider
}

func NewResourceExample() resource.Resource {
    return &resourceExample{}
}

func (r *resourceExample) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_example"
}

func (r *resourceExample) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages an Example resource",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "The unique identifier of the resource",
            },
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "The name of the resource",
            },
        },
    }
}

func (r *resourceExample) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    r.provider = req.ProviderData.(*Provider)
}

func (r *resourceExample) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Implementation
}

func (r *resourceExample) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // Implementation
}

func (r *resourceExample) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Implementation
}

func (r *resourceExample) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // Implementation
}

func (r *resourceExample) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

## Schema Patterns

### Required vs Optional vs Computed

```go
Attributes: map[string]schema.Attribute{
    // Required field - user must provide
    "name": schema.StringAttribute{
        Required:            true,
        MarkdownDescription: "The name of the resource",
    },

    // Optional field - user may provide, has default
    "description": schema.StringAttribute{
        Optional:            true,
        MarkdownDescription: "Optional description",
    },

    // Computed field - set by API, read-only
    "id": schema.StringAttribute{
        Computed:            true,
        MarkdownDescription: "The unique identifier",
    },

    // Optional + Computed - user can set or use API default
    "status": schema.StringAttribute{
        Optional:            true,
        Computed:            true,
        MarkdownDescription: "Resource status",
    },
}
```

### Nested Objects

```go
// Single nested object
Attributes: map[string]schema.Attribute{
    "cloud_config": schema.SingleNestedAttribute{
        Required: true,
        Attributes: map[string]schema.Attribute{
            "region": schema.StringAttribute{Required: true},
            "vpc_id": schema.StringAttribute{Optional: true},
        },
    },
}

// List of nested objects
Attributes: map[string]schema.Attribute{
    "machine_pools": schema.ListNestedAttribute{
        Required: true,
        NestedObject: schema.NestedAttributeObject{
            Attributes: map[string]schema.Attribute{
                "name": schema.StringAttribute{Required: true},
                "count": schema.Int64Attribute{Required: true},
                "instance_type": schema.StringAttribute{Required: true},
            },
        },
    },
}
```

### Complex Types

```go
// Map of strings
"tags": schema.MapAttribute{
    Optional:    true,
    ElementType: types.StringType,
},

// List of strings
"azs": schema.ListAttribute{
    Required:    true,
    ElementType: types.StringType,
},

// Set of strings (no duplicates)
"security_groups": schema.SetAttribute{
    Optional:    true,
    ElementType: types.StringType,
},
```

## State Management Patterns

### Flattening API Response

```go
// API returns complex structure
type APICluster struct {
    UID          string
    Name         string
    MachinePools []APIMachinePool
    Metadata     map[string]string
}

// Terraform model
type ClusterModel struct {
    ID           types.String `tfsdk:"id"`
    Name         types.String `tfsdk:"name"`
    MachinePools types.List   `tfsdk:"machine_pools"`
    Tags         types.Map    `tfsdk:"tags"`
}

// Flattening function
func flattenCluster(cluster *APICluster) *ClusterModel {
    // Convert machine pools
    machinePools, _ := types.ListValueFrom(
        ctx,
        types.ObjectType{AttrTypes: machinePoolAttrTypes},
        flattenMachinePools(cluster.MachinePools),
    )

    // Convert tags
    tags, _ := types.MapValueFrom(
        ctx,
        types.StringType,
        cluster.Metadata,
    )

    return &ClusterModel{
        ID:           types.StringValue(cluster.UID),
        Name:         types.StringValue(cluster.Name),
        MachinePools: machinePools,
        Tags:         tags,
    }
}
```

### Expanding Terraform Config

```go
// Terraform model to API request
func expandClusterRequest(model *ClusterModel) *APIClusterRequest {
    // Extract machine pools from List
    var machinePools []MachinePoolModel
    model.MachinePools.ElementsAs(ctx, &machinePools, false)

    // Convert to API format
    apiPools := make([]APIMachinePool, len(machinePools))
    for i, pool := range machinePools {
        apiPools[i] = APIMachinePool{
            Name:         pool.Name.ValueString(),
            Count:        int(pool.Count.ValueInt64()),
            InstanceType: pool.InstanceType.ValueString(),
        }
    }

    // Extract tags from Map
    tags := make(map[string]string)
    model.Tags.ElementsAs(ctx, &tags, false)

    return &APIClusterRequest{
        Name:         model.Name.ValueString(),
        MachinePools: apiPools,
        Metadata:     tags,
    }
}
```

## Async Resource Patterns

### Long-Running Operations

```go
func (r *resourceClusterAws) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan ClusterModel
    diags := req.Plan.Get(ctx, &plan)

    // 1. Initiate creation
    cluster, err := r.provider.client.CreateCluster(expandClusterRequest(&plan))
    if err != nil {
        resp.Diagnostics.AddError("Error creating cluster", err.Error())
        return
    }

    // 2. Wait for cluster to be ready
    err = r.waitForClusterReady(ctx, cluster.UID, 30*time.Minute)
    if err != nil {
        resp.Diagnostics.AddError("Error waiting for cluster", err.Error())
        return
    }

    // 3. Read final state
    finalCluster, err := r.provider.client.GetCluster(cluster.UID)
    if err != nil {
        resp.Diagnostics.AddError("Error reading cluster", err.Error())
        return
    }

    // 4. Set state
    state := flattenCluster(finalCluster)
    diags = resp.State.Set(ctx, state)
}

func (r *resourceClusterAws) waitForClusterReady(ctx context.Context, id string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)

    for {
        if time.Now().After(deadline) {
            return fmt.Errorf("timeout waiting for cluster to be ready")
        }

        cluster, err := r.provider.client.GetCluster(id)
        if err != nil {
            return err
        }

        switch cluster.Status {
        case "Running":
            return nil
        case "Failed", "Error":
            return fmt.Errorf("cluster creation failed: %s", cluster.StatusMessage)
        default:
            // Still provisioning, wait and retry
            time.Sleep(30 * time.Second)
        }
    }
}
```

## Update Patterns

### Detecting Changes

```go
func (r *resourceClusterAws) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan, state ClusterModel
    req.Plan.Get(ctx, &plan)
    req.State.Get(ctx, &state)

    // Check what changed
    nameChanged := !plan.Name.Equal(state.Name)
    poolsChanged := !plan.MachinePools.Equal(state.MachinePools)
    tagsChanged := !plan.Tags.Equal(state.Tags)

    if nameChanged {
        // Update name
        err := r.provider.client.UpdateClusterName(state.ID.ValueString(), plan.Name.ValueString())
        if err != nil {
            resp.Diagnostics.AddError("Error updating name", err.Error())
            return
        }
    }

    if poolsChanged {
        // Update machine pools (requires reconciliation)
        err := r.updateMachinePools(ctx, state.ID.ValueString(), plan.MachinePools, state.MachinePools)
        if err != nil {
            resp.Diagnostics.AddError("Error updating machine pools", err.Error())
            return
        }
    }

    // ... handle other changes
}
```

### In-Place vs Recreate Updates

```go
// Mark attributes that require recreation
Attributes: map[string]schema.Attribute{
    "cloud": schema.StringAttribute{
        Required: true,
        PlanModifiers: []planmodifier.String{
            stringplanmodifier.RequiresReplace(), // Changing cloud requires recreate
        },
    },
    "name": schema.StringAttribute{
        Required: true,
        // Can be updated in-place, no plan modifier
    },
}
```

## Error Handling Patterns

### Graceful Error Messages

```go
func (r *resourceClusterAws) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    cluster, err := r.provider.client.CreateCluster(request)
    if err != nil {
        // Check error type and provide helpful message
        if apiErr, ok := err.(*client.APIError); ok {
            switch apiErr.StatusCode {
            case 400:
                resp.Diagnostics.AddError(
                    "Invalid Configuration",
                    fmt.Sprintf("The cluster configuration is invalid: %s", apiErr.Message),
                )
            case 403:
                resp.Diagnostics.AddError(
                    "Permission Denied",
                    "You don't have permission to create clusters. Check your API key permissions.",
                )
            case 409:
                resp.Diagnostics.AddError(
                    "Cluster Already Exists",
                    fmt.Sprintf("A cluster with name '%s' already exists", request.Name),
                )
            default:
                resp.Diagnostics.AddError("Error Creating Cluster", err.Error())
            }
        } else {
            resp.Diagnostics.AddError("Error Creating Cluster", err.Error())
        }
        return
    }
}
```

### Partial State on Failure

```go
func (r *resourceClusterAws) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    cluster, err := r.provider.client.CreateCluster(request)
    if err != nil {
        resp.Diagnostics.AddError("Error creating cluster", err.Error())
        return
    }

    // Set ID immediately so resource can be imported/recovered
    state := &ClusterModel{
        ID: types.StringValue(cluster.UID),
    }
    resp.State.Set(ctx, state)

    // Now wait for ready - if this fails, user can still import
    err = r.waitForClusterReady(ctx, cluster.UID, timeout)
    if err != nil {
        resp.Diagnostics.AddError(
            "Cluster Created But Not Ready",
            fmt.Sprintf("Cluster %s was created but did not become ready: %s. "+
                "You can import this cluster manually or destroy it.", cluster.UID, err.Error()),
        )
        return
    }

    // ... continue with full state
}
```

## Import Patterns

### Simple Import

```go
func (r *resourceClusterAws) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Use cluster UID as import ID
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

### Complex Import with Validation

```go
func (r *resourceClusterAws) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    id := req.ID

    // Validate cluster exists and is accessible
    cluster, err := r.provider.client.GetCluster(id)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Importing Cluster",
            fmt.Sprintf("Could not find cluster with ID %s: %s", id, err.Error()),
        )
        return
    }

    // Verify it's the right type
    if cluster.CloudType != "aws" {
        resp.Diagnostics.AddError(
            "Invalid Cluster Type",
            fmt.Sprintf("Cluster %s is type '%s', expected 'aws'", id, cluster.CloudType),
        )
        return
    }

    // Set ID for subsequent Read
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
```

## Validation Patterns

### Attribute Validators

```go
Attributes: map[string]schema.Attribute{
    "instance_type": schema.StringAttribute{
        Required: true,
        Validators: []validator.String{
            stringvalidator.OneOf("t3.small", "t3.medium", "t3.large"),
        },
    },
    "count": schema.Int64Attribute{
        Required: true,
        Validators: []validator.Int64{
            int64validator.Between(1, 100),
        },
    },
    "email": schema.StringAttribute{
        Optional: true,
        Validators: []validator.String{
            stringvalidator.RegexMatches(
                regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
                "must be a valid email address",
            ),
        },
    },
}
```

### Cross-Attribute Validation

```go
func (r *resourceClusterAws) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
    var config ClusterModel
    diags := req.Config.Get(ctx, &config)

    // Validate high availability requires odd number of master nodes
    if config.HighAvailability.ValueBool() {
        var pools []MachinePoolModel
        config.MachinePools.ElementsAs(ctx, &pools, false)

        for _, pool := range pools {
            if pool.ControlPlane.ValueBool() && pool.Count.ValueInt64()%2 == 0 {
                resp.Diagnostics.AddError(
                    "Invalid Configuration",
                    "High availability mode requires an odd number of control plane nodes",
                )
            }
        }
    }
}
```

## Documentation Generation

### Attributes with Examples

```go
Attributes: map[string]schema.Attribute{
    "name": schema.StringAttribute{
        Required: true,
        MarkdownDescription: "The name of the cluster. Must be unique within the project.\n\n" +
            "Example: `production-cluster-01`",
    },
    "tags": schema.MapAttribute{
        Optional:    true,
        ElementType: types.StringType,
        MarkdownDescription: "Key-value pairs to tag the cluster.\n\n" +
            "Example:\n```hcl\ntags = {\n  environment = \"production\"\n  team = \"platform\"\n}\n```",
    },
}
```
