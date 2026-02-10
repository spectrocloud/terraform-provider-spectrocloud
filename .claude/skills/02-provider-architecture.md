# Terraform Provider Architecture

## Provider Framework

This provider uses the **Terraform Plugin Framework** (not the older SDK v2).

### Provider Structure
```go
type Provider struct {
    client *client.V1Client
    // Provider-level configuration
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
    // Define provider schema (host, api_key, project_uid, etc.)
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // Initialize API client
    // Validate credentials
    // Store client in provider data
}
```

## Resource Implementation Pattern

### Resource Structure
```go
type resourceClusterAws struct {
    provider *Provider
}

func (r *resourceClusterAws) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    // Define resource schema with attributes
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{Computed: true},
            "name": schema.StringAttribute{Required: true},
            "cloud_account_id": schema.StringAttribute{Required: true},
            // ... more attributes
        },
    }
}
```

### CRUD Operations

#### Create
```go
func (r *resourceClusterAws) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // 1. Parse Terraform config from request
    var plan ClusterAwsModel
    diags := req.Plan.Get(ctx, &plan)

    // 2. Build API request from plan
    clusterRequest := r.buildClusterRequest(plan)

    // 3. Call API
    cluster, err := r.provider.client.CreateCluster(clusterRequest)

    // 4. Wait for cluster to be ready (if applicable)
    err = r.waitForCluster(ctx, cluster.UID)

    // 5. Update state with created resource
    state := r.mapClusterToState(cluster)
    diags = resp.State.Set(ctx, &state)
}
```

#### Read
```go
func (r *resourceClusterAws) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // 1. Get current state
    var state ClusterAwsModel
    diags := req.State.Get(ctx, &state)

    // 2. Read from API
    cluster, err := r.provider.client.GetCluster(state.ID.ValueString())
    if err != nil {
        if isNotFoundError(err) {
            resp.State.RemoveResource(ctx) // Resource deleted outside Terraform
            return
        }
        resp.Diagnostics.AddError("Error reading cluster", err.Error())
        return
    }

    // 3. Update state with current values
    newState := r.mapClusterToState(cluster)
    diags = resp.State.Set(ctx, &newState)
}
```

#### Update
```go
func (r *resourceClusterAws) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // 1. Get plan and current state
    var plan, state ClusterAwsModel
    diags := req.Plan.Get(ctx, &plan)
    diags = req.State.Get(ctx, &state)

    // 2. Build update request with only changed fields
    updateRequest := r.buildUpdateRequest(state, plan)

    // 3. Call API
    cluster, err := r.provider.client.UpdateCluster(state.ID.ValueString(), updateRequest)

    // 4. Wait for update to complete
    err = r.waitForCluster(ctx, cluster.UID)

    // 5. Update state
    newState := r.mapClusterToState(cluster)
    diags = resp.State.Set(ctx, &newState)
}
```

#### Delete
```go
func (r *resourceClusterAws) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // 1. Get current state
    var state ClusterAwsModel
    diags := req.State.Get(ctx, &state)

    // 2. Call API to delete
    err := r.provider.client.DeleteCluster(state.ID.ValueString())

    // 3. Wait for deletion to complete
    err = r.waitForClusterDeletion(ctx, state.ID.ValueString())

    // 4. State is automatically removed on successful return
}
```

## Data Source Implementation

### Data Source Structure
```go
type dataSourceCluster struct {
    provider *Provider
}

func (d *dataSourceCluster) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    // 1. Get config (search criteria)
    var config ClusterDataSourceModel
    diags := req.Config.Get(ctx, &config)

    // 2. Query API
    cluster, err := d.provider.client.GetCluster(config.ID.ValueString())

    // 3. Set state with found data
    state := d.mapClusterToState(cluster)
    diags = resp.State.Set(ctx, &state)
}
```

## State Management

### Type Mapping
```go
// Terraform model (what Terraform sees)
type ClusterAwsModel struct {
    ID             types.String `tfsdk:"id"`
    Name           types.String `tfsdk:"name"`
    CloudAccountID types.String `tfsdk:"cloud_account_id"`
    Tags           types.Map    `tfsdk:"tags"`
    // ...
}

// API model (what API returns)
type APICluster struct {
    UID            string
    Name           string
    CloudAccountID string
    Metadata       map[string]string
    // ...
}

// Conversion functions
func (r *resourceClusterAws) mapClusterToState(cluster *APICluster) ClusterAwsModel {
    return ClusterAwsModel{
        ID:             types.StringValue(cluster.UID),
        Name:           types.StringValue(cluster.Name),
        CloudAccountID: types.StringValue(cluster.CloudAccountID),
        // Convert complex types
        Tags:           r.mapToTerraformMap(cluster.Metadata),
    }
}
```

### Nested Objects
```go
// Use object types for nested structures
type MachinePoolModel struct {
    Name         types.String `tfsdk:"name"`
    Count        types.Int64  `tfsdk:"count"`
    InstanceType types.String `tfsdk:"instance_type"`
}

// In parent resource
type ClusterModel struct {
    // ...
    MachinePools types.List `tfsdk:"machine_pools"` // List of MachinePoolModel
}
```

## Async Operations

### Waiting for Resources
```go
func (r *resourceClusterAws) waitForCluster(ctx context.Context, clusterID string) error {
    return retry.RetryContext(ctx, 30*time.Minute, func() *retry.RetryError {
        cluster, err := r.provider.client.GetCluster(clusterID)
        if err != nil {
            return retry.NonRetryableError(err)
        }

        switch cluster.Status {
        case "Running":
            return nil // Success
        case "Failed", "Error":
            return retry.NonRetryableError(fmt.Errorf("cluster failed: %s", cluster.StatusMessage))
        default:
            return retry.RetryableError(fmt.Errorf("cluster not ready, status: %s", cluster.Status))
        }
    })
}
```

## Error Handling

### Diagnostic Patterns
```go
// Add error
resp.Diagnostics.AddError(
    "Error Creating Cluster",
    fmt.Sprintf("Could not create cluster: %s", err.Error()),
)

// Add warning
resp.Diagnostics.AddWarning(
    "Deprecated Field",
    "Field 'legacy_field' is deprecated, use 'new_field' instead",
)

// Check if diagnostics has errors
if resp.Diagnostics.HasError() {
    return
}
```

## Testing Patterns

### Acceptance Test Structure
```go
func TestAccResourceClusterAws_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccResourceClusterAwsConfig_basic(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "name", "test-cluster"),
                    resource.TestCheckResourceAttrSet("spectrocloud_cluster_aws.test", "id"),
                ),
            },
            {
                Config: testAccResourceClusterAwsConfig_updated(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "name", "test-cluster-updated"),
                ),
            },
        },
    })
}
```

## API Client Architecture

### Client Structure
```go
type V1Client struct {
    baseURL    string
    httpClient *http.Client
    auth       *AuthContext
}

func (c *V1Client) CreateCluster(req *ClusterRequest) (*Cluster, error) {
    // Build request
    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequest("POST", c.baseURL+"/v1/clusters", bytes.NewBuffer(body))
    httpReq.Header.Set("Authorization", "Bearer "+c.auth.Token)

    // Execute with retry
    resp, err := c.doWithRetry(httpReq)

    // Parse response
    var cluster Cluster
    json.NewDecoder(resp.Body).Decode(&cluster)
    return &cluster, nil
}

func (c *V1Client) doWithRetry(req *http.Request) (*http.Response, error) {
    // Implement exponential backoff for 429, 500, 503
    // Return errors for 400, 401, 403, 404
}
```

## Best Practices

1. **Always validate required fields** in schema
2. **Use types.StringValue() etc.** for null-safe values
3. **Implement proper retries** for async operations
4. **Handle 404 gracefully** in Read (resource may be deleted)
5. **Use computed fields** for server-generated values
6. **Implement Import** for existing resources
7. **Add acceptance tests** for all CRUD operations
8. **Document all attributes** in schema descriptions
9. **Use context cancellation** for long operations
10. **Log at appropriate levels** (DEBUG, INFO, ERROR)
