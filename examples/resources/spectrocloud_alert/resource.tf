/*
Note - Spectocloud allow upto 2 alert configurations for cluster health per project, which can be provisioned under one
resource. Below are example for provisioning alert. we cannot have multiple email component under same resource in single
project context, Instead we can add n' recipient in 'identifiers' or set alert_all_users to true.
*/

# Sample with one email & one webhook alert configuration.
resource "spectrocloud_alert" "alert_dev0" {
  project = "dev0"
  is_active = true
  component = "ClusterHealth"
  http {
    method = "POST"
    url = "https://openhook.com/put/dev0"
    body = "{ \"text\": \"{{message}}\" }"
    headers = {
      ApiKey = "test--key--dev0"
      tag = "Health"
      source = "spectrocloud"
    }
  }
  email {
    alert_all_users = false
    identifiers = ["abc@spectrocloud.com"]
  }
}

# Sample with only email alert configuration.
resource "spectrocloud_alert" "alert_dev1" {
  project = "dev1"
  is_active = true
  component = "ClusterHealth"
  email {
    alert_all_users = false
    identifiers = ["abc@spectrocloud.com", "cba@spectrocloud.com"]
  }
}

# Sample with 2 webhook alert configuration
resource "spectrocloud_alert" "alert_dev2" {
  project = "dev2"
  is_active = true
  component = "ClusterHealth"
  http {
    method = "POST"
    url = "https://openhook.com/put/dev2"
    body = "{ \"text\": \"{{message}}\" }"
    headers = {
      ApiKey = "test--key--dev0"
      tag = "Health"
      source = "spectrocloud"
    }
  }
  http {
    method = "POST"
    url = "https://openhook.com/post/dev2"
    body = "{ \"text\": \"{{message}}\" }"
    headers = {
      ApiKey = "test--key--dev0"
      tag = "Health"
      source = "spectrocloud"
    }
  }
}

# Sample with only webhook alert configuration.
resource "spectrocloud_alert" "alert_dev3" {
  project = "dev3"
  is_active = true
  component = "ClusterHealth"
  http {
    method = "POST"
    url = "https://openhook.com/put/dev3"
    body = "{ \"text\": \"{{message}}\" }"
    headers = {
      ApiKey = "test--key--dev0"
      tag = "Health"
      source = "spectrocloud"
    }
  }
}