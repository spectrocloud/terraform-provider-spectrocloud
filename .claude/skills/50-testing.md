---
skill: Testing Guide for Terraform Provider Spectrocloud
description: Comprehensive testing strategies for the Terraform provider including unit tests for individual functions, acceptance tests for full resource lifecycle against real API, test configuration helpers, and patterns for testing create/read/update/delete/import operations with proper assertions and cleanup.
type: testing
repository: terraform-provider-spectrocloud
team: tools
topics: [testing, terraform, acceptance, unit, go, integration, api, lifecycle, assertions]
difficulty: intermediate
last_updated: 2026-02-09
related_skills: [02-provider-architecture.md, 03-resource-patterns.md, 01-getting-started.md]
memory_references: []
---

# Testing Guide for Terraform Provider Spectrocloud

## Test Structure

```
terraform-provider-spectrocloud/
├── spectrocloud/
│   ├── resource_cluster_aws_test.go      # Unit tests for AWS cluster resource
│   ├── resource_cluster_aws_acc_test.go  # Acceptance tests for AWS cluster
│   ├── data_source_cluster_test.go       # Data source tests
│   └── provider_test.go                  # Provider-level tests
└── tests/
    └── integration/                       # Integration test scenarios
```

## Test Types

### Unit Tests

Test individual functions and methods without making API calls.

```go
func TestFlattenMachinePools(t *testing.T) {
    input := []*models.MachinePool{
        {
            Name:         "master-pool",
            Count:        3,
            InstanceType: "t3.medium",
        },
    }

    result := flattenMachinePools(input)

    assert.Equal(t, 1, len(result))
    assert.Equal(t, "master-pool", result[0]["name"])
    assert.Equal(t, 3, result[0]["count"])
}
```

**Running unit tests:**
```bash
make test
# or
go test -v ./spectrocloud/...
```

### Acceptance Tests

Test actual resource lifecycle against a real Spectro Cloud API.

```go
func TestAccResourceClusterAws_basic(t *testing.T) {
    resourceName := "spectrocloud_cluster_aws.test"

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckClusterAwsDestroy,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccResourceClusterAwsConfig_basic("test-cluster"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckClusterAwsExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "name", "test-cluster"),
                    resource.TestCheckResourceAttr(resourceName, "cloud", "aws"),
                    resource.TestCheckResourceAttrSet(resourceName, "id"),
                    resource.TestCheckResourceAttrSet(resourceName, "kubeconfig"),
                ),
            },
            // ImportState testing
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateVerifyIgnore: []string{"kubeconfig"}, // Ignore sensitive fields
            },
            // Update and Read testing
            {
                Config: testAccResourceClusterAwsConfig_basic("test-cluster-updated"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckClusterAwsExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "name", "test-cluster-updated"),
                ),
            },
            // Delete testing automatically occurs at end
        },
    })
}
```

**Running acceptance tests:**
```bash
# Set required environment variables
export SPECTROCLOUD_HOST="api.spectrocloud.com"
export SPECTROCLOUD_PROJECT="project-id"
export SPECTROCLOUD_APIKEY="your-api-key"

# Run all acceptance tests
make testacc

# Run specific test
TF_ACC=1 go test -v ./spectrocloud -run TestAccResourceClusterAws_basic

# Run with timeout (for long-running tests)
TF_ACC=1 go test -v ./spectrocloud -run TestAccResourceClusterAws -timeout 60m
```

## Test Configuration Helpers

### Basic Configuration
```go
func testAccResourceClusterAwsConfig_basic(name string) string {
    return fmt.Sprintf(`
resource "spectrocloud_cluster_aws" "test" {
  name               = %[1]q
  cloud_account_id   = "cloud-account-id"
  cloud_config_id    = "config-id"

  machine_pool {
    name              = "master-pool"
    count             = 3
    instance_type     = "t3.medium"
    azs               = ["us-west-2a"]
  }

  machine_pool {
    name              = "worker-pool"
    count             = 2
    instance_type     = "t3.large"
    azs               = ["us-west-2a", "us-west-2b"]
  }
}
`, name)
}
```

### Configuration with Dependencies
```go
func testAccResourceClusterAwsConfig_withProfile(name string) string {
    return fmt.Sprintf(`
resource "spectrocloud_cluster_profile" "test" {
  name        = "%[1]s-profile"
  description = "Test profile"
  type        = "cluster"
  cloud       = "aws"
}

resource "spectrocloud_cluster_aws" "test" {
  name               = %[1]q
  cloud_account_id   = "cloud-account-id"
  cluster_profile_id = spectrocloud_cluster_profile.test.id

  machine_pool {
    name          = "master-pool"
    count         = 1
    instance_type = "t3.medium"
    azs           = ["us-west-2a"]
  }
}
`, name)
}
```

## Check Functions

### Resource Exists Check
```go
func testAccCheckClusterAwsExists(resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs, ok := s.RootModule().Resources[resourceName]
        if !ok {
            return fmt.Errorf("Not found: %s", resourceName)
        }

        if rs.Primary.ID == "" {
            return fmt.Errorf("No ID is set")
        }

        client := testAccProvider.Meta().(*client.V1Client)
        _, err := client.GetCluster(rs.Primary.ID)
        if err != nil {
            return fmt.Errorf("Cluster not found: %s", err)
        }

        return nil
    }
}
```

### Resource Destroyed Check
```go
func testAccCheckClusterAwsDestroy(s *terraform.State) error {
    client := testAccProvider.Meta().(*client.V1Client)

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "spectrocloud_cluster_aws" {
            continue
        }

        _, err := client.GetCluster(rs.Primary.ID)
        if err == nil {
            return fmt.Errorf("Cluster still exists")
        }

        // Verify it's a 404 error
        if !isNotFoundError(err) {
            return fmt.Errorf("Expected 404, got: %s", err)
        }
    }

    return nil
}
```

## Test Fixtures

### Provider Configuration
```go
const testAccProviderConfig = `
provider "spectrocloud" {
  host        = "api.spectrocloud.com"
  project_uid = "test-project"
  api_key     = "test-key"
}
`

func testAccPreCheck(t *testing.T) {
    if v := os.Getenv("SPECTROCLOUD_HOST"); v == "" {
        t.Fatal("SPECTROCLOUD_HOST must be set for acceptance tests")
    }
    if v := os.Getenv("SPECTROCLOUD_APIKEY"); v == "" {
        t.Fatal("SPECTROCLOUD_APIKEY must be set for acceptance tests")
    }
}
```

### Test Data
```go
var testCloudAccountID = "test-cloud-account-123"
var testClusterProfileID = "test-profile-456"
var testRegion = "us-west-2"
```

## Mocking and Fixtures

### API Client Mock
```go
type mockClient struct {
    getClusters func(string) (*models.Cluster, error)
    createCluster func(*models.ClusterRequest) (*models.Cluster, error)
}

func (m *mockClient) GetCluster(id string) (*models.Cluster, error) {
    if m.getClusters != nil {
        return m.getClusters(id)
    }
    return &models.Cluster{UID: id, Name: "test-cluster"}, nil
}
```

### Test Fixtures
```go
func testClusterFixture() *models.Cluster {
    return &models.Cluster{
        UID:  "cluster-123",
        Name: "test-cluster",
        Status: "Running",
        MachinePools: []*models.MachinePool{
            {
                Name:         "master-pool",
                Count:        3,
                InstanceType: "t3.medium",
            },
        },
    }
}
```

## Test Patterns

### Testing Complex Nested Structures
```go
func TestAccResourceClusterAws_machinePools(t *testing.T) {
    resource.Test(t, resource.TestCase{
        // ... standard setup ...
        Steps: []resource.TestStep{
            {
                Config: testAccResourceClusterAwsConfig_multiplePools(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "machine_pool.#", "2"),
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "machine_pool.0.name", "master-pool"),
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "machine_pool.0.count", "3"),
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "machine_pool.1.name", "worker-pool"),
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "machine_pool.1.count", "5"),
                ),
            },
        },
    })
}
```

### Testing Optional Fields
```go
func TestAccResourceClusterAws_optionalFields(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config: testAccResourceClusterAwsConfig_minimal(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "name", "test"),
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "tags.%", "0"),
                ),
            },
            {
                Config: testAccResourceClusterAwsConfig_withTags(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "tags.%", "2"),
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "tags.env", "test"),
                    resource.TestCheckResourceAttr("spectrocloud_cluster_aws.test", "tags.team", "platform"),
                ),
            },
        },
    })
}
```

### Testing Error Cases
```go
func TestAccResourceClusterAws_invalidConfig(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config:      testAccResourceClusterAwsConfig_invalid(),
                ExpectError: regexp.MustCompile("invalid instance type"),
            },
        },
    })
}
```

## CI/CD Integration

### GitHub Actions Workflow
```yaml
name: Acceptance Tests

on:
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Unit Tests
        run: make test

      - name: Run Acceptance Tests
        env:
          SPECTROCLOUD_HOST: ${{ secrets.SPECTROCLOUD_HOST }}
          SPECTROCLOUD_APIKEY: ${{ secrets.SPECTROCLOUD_APIKEY }}
          SPECTROCLOUD_PROJECT: ${{ secrets.SPECTROCLOUD_PROJECT }}
        run: make testacc
```

## Coverage

### Generate Coverage Report
```bash
# Unit test coverage
go test -coverprofile=coverage.out ./spectrocloud/...
go tool cover -html=coverage.out -o coverage.html

# View coverage in terminal
go tool cover -func=coverage.out
```

### Coverage Requirements
- **Minimum**: 70% code coverage for new resources
- **Target**: 80%+ coverage for all code
- **Critical paths**: 100% coverage (authentication, API calls)

## Test Best Practices

1. **Clean up resources**: Always implement CheckDestroy
2. **Use random names**: Avoid test conflicts with `acctest.RandomWithPrefix()`
3. **Test all CRUD operations**: Create, Read, Update, Delete, Import
4. **Test error cases**: Invalid configs, API failures
5. **Use fixtures**: Reuse test data structures
6. **Mock external dependencies**: Use mocks for unit tests
7. **Set appropriate timeouts**: Cluster provisioning can take 20+ minutes
8. **Test optional fields**: Verify defaults and updates
9. **Test computed fields**: Ensure read-only fields are populated
10. **Use parallel tests**: Speed up test suites with `t.Parallel()`

## Troubleshooting Tests

### Test Hangs
- Check API rate limits
- Increase timeout: `-timeout 60m`
- Verify cluster creation completes
- Check for infinite retry loops

### Flaky Tests
- Add retry logic for eventual consistency
- Increase wait times for async operations
- Use resource.Retry for transient failures
- Check for race conditions

### Failed Cleanup
- Manually delete leftover resources
- Check test account quotas
- Verify API credentials haven't expired
- Review destroy functions

## Example: Complete Test Suite

```go
func TestAccResourceClusterAws(t *testing.T) {
    t.Run("basic", func(t *testing.T) {
        testAccResourceClusterAws_basic(t)
    })
    t.Run("update", func(t *testing.T) {
        testAccResourceClusterAws_update(t)
    })
    t.Run("import", func(t *testing.T) {
        testAccResourceClusterAws_import(t)
    })
    t.Run("tags", func(t *testing.T) {
        testAccResourceClusterAws_tags(t)
    })
    t.Run("machinePools", func(t *testing.T) {
        testAccResourceClusterAws_machinePools(t)
    })
}
```
