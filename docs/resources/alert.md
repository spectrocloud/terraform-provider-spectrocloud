---
page_title: "spectrocloud_alert Resource - terraform-provider-spectrocloud"
subcategory: ""
description: Provisioning Cluster Health alerts (email | http).|-
  
---

# Resource `spectrocloud_alert`



## Example Usage
#### Note
- Spectro Cloud creates up to two alerts from the below combinations:
    - 1 email alert (can add any number of email recipient) & 1 http webhook alert.
    - 2 http webhook alert configuration.
    - Any one alert type Email/http.

  (Documentation [Cluster Health Alerts](#https://docs.spectrocloud.com/clusters/cluster-management/health-alerts#overview))




```terraform
Type : HTTP
-----------------------------------------------------------
resource "spectrocloud_alert" "alert_http" {
  project = "Default"
  is_active = true
  component = "ClusterHealth"
  http {
    method  = "POST"
    url     = "https://openhook.com/put/edit2"
    body    = "{ \"text\": \"{{message}}\" }"
    headers = {
      type = "test--key--dev0"
      tag    = "Health"
      source = "spectrocloud"
    }
  }
  type = "http"
  alert_all_users = false
}

Type : EMAIL
-----------------------------------------------------------
resource "spectrocloud_alert" "alert_email" {
  project = "Default"
  is_active = true
  component = "ClusterHealth"
  type = "email"
  identifiers = ["siva@spectrocloud.com", "anand@spectrocloud.com"]
  alert_all_users = false
}
```

## Schema

### Required

- **project (name)** (String)
- **is_active** (Bool)
- **component** (Bool)
- **alert (http/email)** (Map)


### Optional

- **identifiers** (String)
- **body** (String)
- **headers** ([]Map)
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)
