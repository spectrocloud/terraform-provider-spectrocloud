 resource "spectrocloud_application_profile" "mysql_profile" {
  name        = "tf_test_app_profile"
  description = "Application Profile with MySQL 5.7 database"
  version     = "1.0.0"
  context     = "project"
  cloud       = "all"

  pack {
    name            = "mysql-operator-1"  # Pack name
    type            = "operator-instance"
    source_app_tier = "65555c69bca922f250d7c4ff"  # Get this from UI or existing profile
    properties = {
      "dbRootPassword" = base64encode("Testmanimaran")
      "dbVolumeSize"   = "15"
      "version"        = "5.7"
    }
    tag = "1.0.0"
  }

  tags = ["database:mysql", "version:5.7"]
}