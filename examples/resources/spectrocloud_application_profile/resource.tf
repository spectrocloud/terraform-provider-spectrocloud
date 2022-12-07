resource "spectrocloud_application_profile" "app_profile_all_tiers" {
  name        = "profile-all-tiers-test"
  version     = "1.0.0"
  context     = "project"
  tags        = ["sivaa", "terraform"]
  description = "test"
  cloud       = "all"
  # Sample Container Tier
  pack {
    name            = "container-tier"
    type            = "container"
    registry_uid    = "63245b43e52f17068237e0cd"
    source_app_tier = "636223d44ead10f3bdff62d9"
    values          = "pack:\n  namespace: \"{{.spectro.system.appdeployment.tiername}}-ns\"\n  releaseNameOverride: \"{{.spectro.system.appdeployment.tiername}}\"\npostReadinessHooks:\n  outputParameters:\n    - name: CONTAINER_SVC\n      type: lookupSecret\n      spec:\n        namespace: \"{{.spectro.system.appdeployment.tiername}}-ns\"\n        secretName: \"{{.spectro.system.appdeployment.tiername}}-custom-secret\"\n        ownerReference:\n          apiVersion: v1\n          kind: Service\n          name: \"{{.spectro.system.appdeployment.tiername}}-svc\"\n        keyToCheck: metadata.name\n    - name: CONTAINER_SVC_PORT\n      type: lookupSecret\n      spec:\n        namespace: \"{{.spectro.system.appdeployment.tiername}}-ns\"\n        secretName: \"{{.spectro.system.appdeployment.tiername}}-custom-secret\"\n        ownerReference:\n          apiVersion: v1\n          kind: Service\n          name: \"{{.spectro.system.appdeployment.tiername}}-svc\"\n        keyToCheck: spec.ports[0].port\n        keyFormat: string, number\ncontainerService:\n  serviceName: \"{{.spectro.system.appdeployment.tiername}}-svc\"\n  registryUrl: \"\"\n  image: alphine\n  access: public\n  ports:\n    - \"8080\"\n  serviceType: LoadBalancer\n  args:\n    - $TEST\n  command:\n    - sh\n    - ./start.sh\n  env:\n    - name: TEST\n      value: \"true\"\n  volumeName: TestVolume\n  volumeSize: 10\n  pathToMount: /pack\n"
  }
  # Sample Helm Tier
  pack {
    name            = "kafka-tier"
    type            = "helm"
    registry_uid    = "605f66e748cd5ef5bd070ba8"
    source_app_tier = "638f21ac7f63fb9677625873"
    manifest {
      name    = "test"
      content = "annotations:\n  category: Infrastructure\napiVersion: v2\nappVersion: 3.3.1\ndependencies:\n  - condition: zookeeper.enabled\n    name: zookeeper\n    repository: https://charts.bitnami.com/bitnami\n    version: 11.x.x\n  - name: common\n    repository: https://charts.bitnami.com/bitnami\n    tags:\n      - bitnami-common\n    version: 2.x.x\ndescription: Apache Kafka is a distributed streaming platform designed to build real-time pipelines and can be used as a message broker or as a replacement for a log aggregation solution for big data applications.\nengine: gotpl\nhome: https://github.com/bitnami/charts/tree/main/bitnami/kafka\nicon: https://bitnami.com/assets/stacks/kafka/img/kafka-stack-220x234.png\nkeywords:\n  - kafka\n  - zookeeper\n  - streaming\n  - producer\n  - consumer\nmaintainers:\n  - name: Bitnami\n    url: https://github.com/bitnami/charts\nname: kafka\nsources:\n  - https://github.com/bitnami/containers/tree/main/bitnami/kafka\n  - https://kafka.apache.org/\nversion: 20.0.0"
    }
  }
  # Sample Manifest Tier
  pack {
    name          = "manifest-3"
    type          = "manifest"
    install_order = 0
    manifest {
      name    = "test-manifest-3"
      content = "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx-deployment\n  labels:\n    app: nginx\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.14.2\n        ports:\n        - containerPort: 80"
    }
  }
  # Sample DB Tier's
  pack {
    type            = "operator-instance"
    name            = "minio-operator-stage"
    source_app_tier = "6384db506a675d8599aa95f5"
    properties = {
      "minioUsername"     = "miniostaging"
      "minioUserPassword" = base64encode("test123!wewe!")
      "volumeSize"        = "10"
    }
  }
  pack {
    type            = "operator-instance"
    name            = "mysql-3-stage"
    source_app_tier = "636c0714c196e565df7a7b37"
    properties = {
      "dbRootPassword" = base64encode("test123!wewe!")
      "dbVolumeSize"   = "20"
      "dbVersion"      = "5.7"
    }
  }
  pack {
    type            = "operator-instance"
    name            = "redis-4-stage"
    source_app_tier = "637d7ef64e3ddd9df17ae2b9"
    properties = {
      "databaseName"       = "redsitstaging"
      "databaseVolumeSize" = "10"
    }
  }
}
