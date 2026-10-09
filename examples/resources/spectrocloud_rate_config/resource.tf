resource "spectrocloud_rate_config" "rate_config" {
  # Public clouds split the instance price between compute and memory.
  # Each pair of proportions must add up to 100.
  aws {
    compute_optimized {
      compute_rate_proportion = 65
      memory_rate_proportion  = 35
    }
    memory_optimized {
      compute_rate_proportion = 25
      memory_rate_proportion  = 75
    }
  }

  azure {
    compute_optimized {
      compute_rate_proportion = 65
      memory_rate_proportion  = 35
    }
    memory_optimized {
      compute_rate_proportion = 25
      memory_rate_proportion  = 75
    }
  }

  # Private clouds are priced per resource unit, in US dollars.
  vsphere {
    cpu_unit_price_per_hour         = 0.025
    gpu_unit_price_per_hour         = 3.100
    memory_unit_price_gib_per_hour  = 0.003
    storage_unit_price_gib_per_hour = 0.00003
  }

  maas {
    cpu_unit_price_per_hour         = 0.021811
    gpu_unit_price_per_hour         = 2.933908
    memory_unit_price_gib_per_hour  = 0.002923
    storage_unit_price_gib_per_hour = 0.000028
  }

  # One block per custom cloud registered in the tenant.
  custom {
    cloud_type = "my-custom-cloud"
    rate_config {
      cpu_unit_price_per_hour         = 0.030
      gpu_unit_price_per_hour         = 3.500
      memory_unit_price_gib_per_hour  = 0.004
      storage_unit_price_gib_per_hour = 0.00004
    }
  }
}

## import the existing tenant rate config
#import {
#  to = spectrocloud_rate_config.rate_config
#  id = "{tenantUID}" // tenant-uid or organization name
#}
