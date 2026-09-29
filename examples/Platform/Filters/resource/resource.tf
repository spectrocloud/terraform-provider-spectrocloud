# Defines a reusable tag-based filter (e.g. to scope which clusters a cluster group or RBAC
# binding applies to). Day-2 mutability: nothing on this resource is ForceNew - metadata.name,
# spec, and all nested filter_group/filters attributes update in place.
resource "spectrocloud_filter" "example" {
  # metadata:
  #   name - Required. The filter's display name.
  metadata {
    name = "resourcefilter2"
  }

  spec {
    # filter_group (Required, exactly one):
    #   conjunction - Required. How the filters below combine. Allowed: "and", "or".
    filter_group {
      conjunction = "and"

      # filters (Required, one or more - each block is one condition; testtag1 condition):
      #   key      - Required. The tag key to match against.
      #   negation - Optional, default false. If true, inverts the match (i.e. "not equal").
      #   operator - Required. Comparison operator. Currently only "eq" (equals) is supported.
      #   values   - Required. Values to compare the tag's value against.
      filters {
        key      = "testtag1"
        negation = false
        operator = "eq"
        values   = ["spectro__tag"]
      }

      # filters (testtag2 condition, negated):
      #   key      - Required. The tag key to match against.
      #   negation - Optional, default false. If true, inverts the match (i.e. "not equal").
      #   operator - Required. Comparison operator. Currently only "eq" (equals) is supported.
      #   values   - Required. Values to compare the tag's value against.
      filters {
        key      = "testtag2"
        negation = true
        operator = "eq"
        values   = ["spectro__tag"]
      }
    }
  }
}
