# Apache CloudStack Cluster - Override Scaling Guide

This guide explains the `override_scaling` feature for Apache CloudStack clusters, which provides fine-grained control over rolling update behavior.

## Table of Contents

- [Overview](#overview)
- [Update Strategy Types](#update-strategy-types)
- [Override Scaling Configuration](#override-scaling-configuration)
- [Common Use Cases](#common-use-cases)
- [Examples](#examples)
- [Validation Rules](#validation-rules)
- [Best Practices](#best-practices)

## Overview

The `override_scaling` feature allows you to customize how Kubernetes rolling updates behave for your machine pools. It provides control over:

- **max_surge**: Maximum number of nodes that can be created above the desired count during an update
- **max_unavailable**: Maximum number of nodes that can be unavailable during an update

This gives you precise control over the trade-off between update speed and service availability.

## Update Strategy Types

Apache CloudStack clusters support three update strategies:

### 1. RollingUpdateScaleOut (Default)
- Adds new nodes before removing old ones
- Ensures capacity is maintained or increased during updates
- Slower but safer for production workloads

```hcl
machine_pool {
  # ...
  update_strategy = "RollingUpdateScaleOut"
}
```

### 2. RollingUpdateScaleIn
- Removes old nodes before adding new ones
- Reduces capacity temporarily during updates
- Faster but may impact service availability

```hcl
machine_pool {
  # ...
  update_strategy = "RollingUpdateScaleIn"
}
```

### 3. OverrideScaling
- Custom control with `max_surge` and `max_unavailable` parameters
- Provides the most flexibility for fine-tuning update behavior
- **Requires** the `override_scaling` block to be specified

```hcl
machine_pool {
  # ...
  update_strategy = "OverrideScaling"
  
  override_scaling {
    max_surge       = "1"
    max_unavailable = "0"
  }
}
```

## Override Scaling Configuration

### Syntax

```hcl
override_scaling {
  max_surge       = "<value>"  # Required: Number or percentage
  max_unavailable = "<value>"  # Required: Number or percentage
}
```

### Value Format

Values can be specified as:

1. **Absolute Numbers**: `"0"`, `"1"`, `"2"`, `"3"`, etc.
   - Specifies exact number of nodes

2. **Percentages**: `"10%"`, `"25%"`, `"50%"`, `"100%"`, etc.
   - Calculated based on the desired node count
   - Rounded up for fractional values

### Parameters

#### max_surge
- Maximum number of nodes that can be created **above** the desired count
- Higher values = faster updates but more temporary resource usage
- Can be `"0"` to prevent any surge capacity

#### max_unavailable
- Maximum number of nodes that can be **unavailable** during updates
- Lower values = higher availability but slower updates
- Can be `"0"` for zero-downtime updates (requires max_surge > 0)

## Common Use Cases

### Zero Downtime Updates (Production)

**Goal**: Maintain full capacity at all times, no service disruption

```hcl
machine_pool {
  name            = "production-workers"
  count           = 3
  update_strategy = "OverrideScaling"
  
  override_scaling {
    max_surge       = "1"    # Add 1 new node at a time
    max_unavailable = "0"    # Never reduce capacity
  }
}
```

**Behavior**: Creates a new node first, waits for it to be ready, then removes an old node. Repeats until all nodes are updated.

### Balanced Updates (Staging)

**Goal**: Balance between update speed and availability

```hcl
machine_pool {
  name            = "staging-workers"
  count           = 4
  update_strategy = "OverrideScaling"
  
  override_scaling {
    max_surge       = "25%"   # Can add 1 node (25% of 4)
    max_unavailable = "25%"   # Can have 1 node down (25% of 4)
  }
}
```

**Behavior**: Can update multiple nodes in parallel, with some temporary over-capacity and brief under-capacity periods.

### Fast Updates (Development)

**Goal**: Update as quickly as possible, availability is less critical

```hcl
machine_pool {
  name            = "dev-workers"
  count           = 3
  update_strategy = "OverrideScaling"
  
  override_scaling {
    max_surge       = "2"    # Can add 2 nodes
    max_unavailable = "1"    # 1 node can be down
  }
}
```

**Behavior**: Aggressively updates nodes in parallel, trading availability for speed.

### Cost-Optimized Updates

**Goal**: Minimize temporary resource costs during updates

```hcl
machine_pool {
  name            = "cost-optimized-workers"
  count           = 5
  update_strategy = "OverrideScaling"
  
  override_scaling {
    max_surge       = "0"    # Never create extra nodes
    max_unavailable = "1"    # Remove one at a time
  }
}
```

**Behavior**: Removes old nodes one at a time and creates replacements. No surge capacity means lower costs but reduced availability during updates.

## Examples

### Example 1: Large Cluster with Percentage-Based Scaling

```hcl
machine_pool {
  name  = "large-worker-pool"
  count = 20
  min   = 10
  max   = 30
  
  placement {
    zone         = "zone-1"
    compute      = "medium-compute"
    network_name = "cluster-network"
  }
  
  update_strategy = "OverrideScaling"
  
  # With count=20:
  # max_surge=10% → 2 nodes
  # max_unavailable=5% → 1 node
  override_scaling {
    max_surge       = "10%"
    max_unavailable = "5%"
  }
  
  additional_labels = {
    "environment" = "production"
    "pool-size"   = "large"
  }
}
```

### Example 2: Multi-Pool Strategy

```hcl
resource "spectrocloud_cluster_apache_cloudstack" "cluster" {
  name = "multi-strategy-cluster"
  # ... cloud_config and cluster_profile ...
  
  # Control plane: Conservative updates
  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "control-plane"
    count                   = 3
    
    # ... placement ...
    
    update_strategy = "RollingUpdateScaleOut"  # Default, safe strategy
  }
  
  # Critical workers: Zero downtime
  machine_pool {
    name  = "critical-workers"
    count = 5
    
    # ... placement ...
    
    update_strategy = "OverrideScaling"
    override_scaling {
      max_surge       = "1"
      max_unavailable = "0"
    }
    
    additional_labels = {
      "workload" = "critical"
    }
  }
  
  # Batch workers: Fast updates
  machine_pool {
    name  = "batch-workers"
    count = 10
    
    # ... placement ...
    
    update_strategy = "OverrideScaling"
    override_scaling {
      max_surge       = "50%"  # 5 nodes
      max_unavailable = "30%"  # 3 nodes
    }
    
    additional_labels = {
      "workload" = "batch"
    }
  }
}
```

## Validation Rules

The following validation rules are enforced:

1. **Required override_scaling block**
   ```hcl
   # ❌ INVALID - Missing override_scaling
   machine_pool {
     update_strategy = "OverrideScaling"
     # ERROR: override_scaling must be specified
   }
   ```

2. **Both parameters required**
   ```hcl
   # ❌ INVALID - Missing max_unavailable
   override_scaling {
     max_surge = "1"
     # ERROR: max_unavailable must be specified
   }
   ```

3. **Valid value format**
   ```hcl
   # ✅ VALID - Absolute numbers
   override_scaling {
     max_surge       = "2"
     max_unavailable = "1"
   }
   
   # ✅ VALID - Percentages
   override_scaling {
     max_surge       = "25%"
     max_unavailable = "10%"
   }
   
   # ❌ INVALID - Empty values
   override_scaling {
     max_surge       = ""  # ERROR: must not be empty
     max_unavailable = ""  # ERROR: must not be empty
   }
   ```

4. **Cannot use override_scaling without OverrideScaling strategy**
   ```hcl
   # ❌ INVALID - Strategy mismatch
   machine_pool {
     update_strategy = "RollingUpdateScaleOut"
     override_scaling {  # ERROR: only valid with OverrideScaling strategy
       max_surge       = "1"
       max_unavailable = "0"
     }
   }
   ```

## Best Practices

### 1. Choose Strategy Based on Environment

- **Production**: Use `OverrideScaling` with `max_unavailable = "0"` for zero downtime
- **Staging**: Use `OverrideScaling` with balanced percentages (e.g., 25%/25%)
- **Development**: Use `RollingUpdateScaleIn` or aggressive `OverrideScaling`

### 2. Consider Cluster Size

- **Small clusters (< 5 nodes)**: Use absolute numbers for predictable behavior
- **Large clusters (> 10 nodes)**: Use percentages for proportional scaling

### 3. Monitor Resource Usage

- Higher `max_surge` values create temporary over-capacity
- Monitor costs and resource limits when using large surge values
- Consider using `max_surge = "0"` for cost-sensitive environments

### 4. Test Update Strategies

- Test your update strategy in non-production environments first
- Measure actual update times and service impact
- Adjust `max_surge` and `max_unavailable` based on observed behavior

### 5. Document Your Choices

```hcl
machine_pool {
  name = "workers"
  count = 5
  
  update_strategy = "OverrideScaling"
  
  # Zero-downtime updates for user-facing services
  # Chosen because we handle 1000+ requests/minute
  # and cannot tolerate any capacity reduction
  override_scaling {
    max_surge       = "1"
    max_unavailable = "0"
  }
  
  additional_labels = {
    "update-strategy-reason" = "high-traffic-zero-downtime"
  }
}
```

### 6. Use Appropriate Node Repave Intervals

Combine `override_scaling` with `node_repave_interval` for regular updates:

```hcl
machine_pool {
  # ...
  update_strategy = "OverrideScaling"
  override_scaling {
    max_surge       = "1"
    max_unavailable = "0"
  }
  
  # Repave nodes every 90 days
  node_repave_interval = 90
}
```

## Troubleshooting

### Updates Taking Too Long

**Problem**: Updates are very slow

**Solutions**:
- Increase `max_surge` to allow more parallel updates
- Increase `max_unavailable` if some downtime is acceptable
- Check if percentage values are too conservative for your cluster size

### Insufficient Resources During Updates

**Problem**: Cluster runs out of resources during updates

**Solutions**:
- Reduce `max_surge` to limit temporary over-capacity
- Increase cluster `max` limit to allow surge capacity
- Consider using `max_surge = "0"` with higher `max_unavailable`

### Service Disruptions During Updates

**Problem**: Users experience downtime during updates

**Solutions**:
- Set `max_unavailable = "0"` for zero-downtime updates
- Increase `max_surge` to maintain more excess capacity
- Ensure application has proper pod disruption budgets

### Validation Errors

**Problem**: Terraform validation fails

**Solutions**:
- Ensure `override_scaling` block is present when using `OverrideScaling` strategy
- Verify both `max_surge` and `max_unavailable` are specified
- Check that values are non-empty strings
- Confirm values are valid numbers or percentages

## Additional Resources

- [Kubernetes Rolling Update Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment)
- [Spectro Cloud Documentation](https://docs.spectrocloud.com/)
- [Terraform Provider Documentation](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs)

## Support

For issues or questions:
1. Check the validation rules and error messages
2. Review the examples in this guide
3. Consult the [Spectro Cloud documentation](https://docs.spectrocloud.com/)
4. Contact Spectro Cloud support

