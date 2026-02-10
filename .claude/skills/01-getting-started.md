---
skill: Getting Started with Terraform Provider Spectrocloud
type: onboarding
repository: terraform-provider-spectrocloud
team: tools
topics: [terraform, provider, setup, development, spectrocloud, iac]
difficulty: beginner
last_updated: 2026-02-09
related_skills: [02-provider-architecture.md, 03-resource-patterns.md]
memory_references: []
---

# Getting Started with Terraform Provider Spectrocloud

## Overview
Terraform provider for managing Spectro Cloud resources. Enables Infrastructure as Code (IaC) for cluster provisioning, profiles, projects, and other Spectro Cloud resources.

## Repository Structure
```
terraform-provider-spectrocloud/
├── .claude/              # Agent configurations and skills
├── spectrocloud/         # Provider implementation
│   ├── resource_*.go     # Resource implementations
│   ├── data_source_*.go  # Data source implementations
│   └── provider.go       # Provider configuration
├── docs/                 # Terraform documentation
├── examples/             # Example configurations
└── tests/                # Acceptance tests
```

## Development Setup

### Prerequisites
- Go 1.21+
- Terraform 1.0+
- Spectro Cloud account with API credentials
- Make

### Building the Provider
```bash
make build
```

### Installing Locally
```bash
make install
```

This installs the provider to:
```
~/.terraform.d/plugins/
```

### Running Tests
```bash
# Unit tests
make test

# Acceptance tests (requires API credentials)
export SPECTROCLOUD_HOST="api.spectrocloud.com"
export SPECTROCLOUD_PROJECT="your-project-id"
export SPECTROCLOUD_APIKEY="your-api-key"
make testacc
```

## Key Concepts

### Provider Configuration
```hcl
provider "spectrocloud" {
  host        = "api.spectrocloud.com"
  project_uid = "your-project-id"
  api_key     = "your-api-key"
}
```

### Resource Types
- **Clusters**: `spectrocloud_cluster_*` - Manage cluster lifecycle
- **Profiles**: `spectrocloud_cluster_profile` - Define cluster configurations
- **Projects**: `spectrocloud_project` - Organize resources
- **Workspaces**: `spectrocloud_workspace` - Virtual workspace management
- **Backups**: `spectrocloud_backup_storage_location` - Backup configurations

### Data Sources
- `spectrocloud_cluster` - Query cluster information
- `spectrocloud_cloudaccount_*` - Cloud account details
- `spectrocloud_registry` - OCI registry information
- `spectrocloud_pack` - Pack information

## Architecture

### Provider Structure
```
Provider Registration
├── Schema Definition (provider.go)
├── Client Configuration
└── Resource/DataSource Registration

Resources
├── CRUD Operations (Create, Read, Update, Delete)
├── Import Support
└── State Management

Data Sources
├── Read Operations
└── Filtering/Querying
```

### API Integration
- Uses Spectro Cloud REST API
- SDK client in `spectrocloud/client/`
- Handles authentication, retries, and rate limiting

## Common Workflows

### Adding a New Resource
1. Define schema in `resource_*.go`
2. Implement CRUD functions
3. Add acceptance tests
4. Generate documentation
5. Update examples

### Updating API Client
1. Modify client in `spectrocloud/client/`
2. Update resource implementations
3. Run tests
4. Update documentation

## Related Repositories
- **palette-cli**: CLI tool for Spectro Cloud
- **hapi**: Spectro Cloud REST API (consumed by this provider)

## Team
**Tools Team** - Responsible for CLI and Terraform provider
