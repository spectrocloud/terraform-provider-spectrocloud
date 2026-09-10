# Defines a reusable tag-based filter (e.g. to scope which clusters a cluster group or RBAC
# binding applies to). Day-2 mutability: nothing on this resource is ForceNew - metadata.name,
# spec, and all nested filter_group/filters attributes update in place.
resource "spectrocloud_filter" "example" {
  metadata {
    # Required. The filter's display name.
    name = "resourcefilter2"
  }

  spec {
    # Required, exactly one filter_group block.
    filter_group {
      # Required. How the filters below combine. Allowed: "and", "or".
      conjunction = "and"

      # Required, one or more. Each filters block is one condition.
      filters {
        # Required. The tag key to match against.
        key = "testtag1"
        # Optional, default false. If true, inverts the match (i.e. "not equal").
        negation = false
        # Required. Comparison operator. Currently only "eq" (equals) is supported.
        operator = "eq"
        # Required. Values to compare the tag's value against.
        values = ["spectro__tag"]
      }

      filters {
        key      = "testtag2"
        negation = true
        operator = "eq"
        values   = ["spectro__tag"]
      }
    }
  }
}
