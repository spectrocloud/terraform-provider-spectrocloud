resource "spectrocloud_macros" "project_macro" {
  macros ={
    "project1" = "val1",
#    "project2" = "val2",
#    "project3" = "val3",
#    "project4" = "val4",
#    "project5" = "val5",
#    "project6" = "val6",
#    "project7" = "val7",
#    "project8" = "val8",
#    "project9" = "val9",
  }
  project = "Default"
}

resource "spectrocloud_macros" "tenant_macro" {
  macros = {
    "tenant1" = "tenant_val1",
#    "tenant2" = "tenant_val2",
#    "tenant3" = "tenant_val3",
#    "tenant4" = "tenant_val4",
#    "tenant5" = "tenant_val5",
  }
}