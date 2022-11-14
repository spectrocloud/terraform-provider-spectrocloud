/*
Note - We can set up to two alerts for cluster health per project. Web-hook can be configured in HTTP component and for
email, we can add a target email recipients or enable alerts for all users in the corresponding project
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