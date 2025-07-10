resource "spectrocloud_macros" "project_macro" {
  macros = {
    "project_macro_1" = "val1",
    "project_macro_2" = "val2",
  }
  context = "project"
}

resource "spectrocloud_macros" "tenant_macro" {
  macros = {
    "tenant_macro_1" = "tenant_val1",
    "tenant_macro_2" = "tenant_val2",
  }
}


#Sample id id = "63d48658fsga0c92a6f230112:tenant"
#import {
#  to = spectrocloud_macros.imported_macros_tenant
#  id = "63d48062b3a0c92a6f230112:tenant"
#}
#import {
#  to = spectrocloud_macros.imported_macros_project
#  id = "67a8e0e3dc76532bf3d8af3c:project"
#}

