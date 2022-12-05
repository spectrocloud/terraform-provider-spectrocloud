resource "spectrocloud_application_profile" "test_profile_tf"{
  name = "profile-herbs"
  version = "1.0.0"
  context = "project"
  tags = ["sivaa", "terraform"]
  description = "test"
  cloud = "all"
  pack {
    type = "operator-instance"
    name = "mysql-3-stage"
    source_app_tier="636c0714c196e565df7a7b37"
    properties = {
      "dbRootPassword" = base64encode("test123!wewe!")
      "dbVolumeSize" = "20"
      "dbVersion" = "5.7"
    }
  }
  pack {
    type = "operator-instance"
    name = "redis-4-stage"
    source_app_tier="637d7ef64e3ddd9df17ae2b9"
    properties = {
      "databaseName" = "redsitstaging"
      "databaseVolumeSize" = "10"
    }
  }
  pack {
    type = "operator-instance"
    name = "minio-operator-stage"
    source_app_tier="6384db506a675d8599aa95f5"
    properties = {
      "minioUsername" = "miniostaging"
      "minioUserPassword" = base64encode("test123!wewe!")
      "volumeSize" = "10"
    }
  }
}
