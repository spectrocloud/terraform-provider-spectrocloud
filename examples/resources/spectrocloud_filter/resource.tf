resource "spectrocloud_filter" "example" {
  metadata {
    name = "resourcefilter2"
  }

  spec {
    filter_group {
      conjunction = "and"

      filters {
        key = "testtag1"
        negation = false
        operator = "eq"
        values = ["spectro__tag"]
      }

      filters {
        key = "testtag2"
        negation = true
        operator = "eq"
        values = ["spectro__tag"]
      }
    }
  }
}
