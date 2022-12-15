data "spectrocloud_registry" "common_registry" {
  name = "Public Repo"
}

data "spectrocloud_registry" "container_registry" {
  name = "automation-pack-registry"
}

data "spectrocloud_registry" "db_registry" {
  name = "svtest"
}

data "spectrocloud_registry" "Bitnami_registry" {
  name = "Bitnami"

}

data "spectrocloud_pack_simple" "redis_pack" {
  type         = "operator-instance"
  name         = "redis-operator"
  version      = "6.2.12-1"
  registry_uid = data.spectrocloud_registry.common_registry.id
}

data "spectrocloud_pack_simple" "mysql_pack" {
  type         = "operator-instance"
  name         = "mysql-operator"
  version      = "0.6.2"
  registry_uid = data.spectrocloud_registry.db_registry.id
}

data "spectrocloud_pack_simple" "minio_pack" {
  type         = "operator-instance"
  name         = "minio-operator"
  version      = "4.5.4"
  registry_uid = data.spectrocloud_registry.db_registry.id
}

data "spectrocloud_pack_simple" "container_pack" {
  type         = "container"
  name         = "container"
  version      = "1.0.0"
  registry_uid = data.spectrocloud_registry.container_registry.id
}

data "spectrocloud_pack_simple" "kafka_pack" {
  type         = "helm"
  name         = "kafka"
  version      = "20.0.0"
  registry_uid = data.spectrocloud_registry.Bitnami_registry.id
}

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
    type            = data.spectrocloud_pack_simple.container_pack.type
    registry_uid    = data.spectrocloud_registry.container_registry.id
    source_app_tier = data.spectrocloud_pack_simple.container_pack.id
    values          = <<-EOT
        containerService:
            serviceName: "{{.spectro.system.appdeployment.tiername}}-svc"
            registryUrl: ""
            image: alphine
            access: public
            ports:
              - "8080"
            serviceType: LoadBalancer
            args:
              - $TEST
            command:
              - sh
              - ./start.sh
            env:
              - name: TEST
                value: "true"
            volumeName: TestVolume
            volumeSize: 10
            pathToMount: /pack/
    EOT
  }
  # Sample Helm Tier
  pack {
    name            = "kafka-tier"
    type            = data.spectrocloud_pack_simple.kafka_pack.type
    registry_uid    = data.spectrocloud_registry.Bitnami_registry.id
    source_app_tier = data.spectrocloud_pack_simple.kafka_pack.id
    manifest {
      name    = "kafka"
      content = <<-EOT
          annotations:
            category: Infrastructure
          apiVersion: v2
          appVersion: 3.3.1
          dependencies:
            - condition: zookeeper.enabled
              name: zookeeper
              repository: https://charts.bitnami.com/bitnami
              version: 11.x.x
            - name: common
              repository: https://charts.bitnami.com/bitnami
              tags:
                - bitnami-common
              version: 2.x.x
          description: Apache Kafka is a distributed streaming platform designed to build real-time pipelines and can be used as a message broker or as a replacement for a log aggregation solution for big data applications.
          engine: gotpl
          home: https://github.com/bitnami/charts/tree/main/bitnami/kafka
          icon: https://bitnami.com/assets/stacks/kafka/img/kafka-stack-220x234.png
          keywords:
            - kafka
            - zookeeper
            - streaming
            - producer
            - consumer
          maintainers:
            - name: Bitnami
              url: https://github.com/bitnami/charts
          name: kafka
          sources:
            - https://github.com/bitnami/containers/tree/main/bitnami/kafka
            - https://kafka.apache.org/
          version: 20.0.0
      EOT
    }
  }
  # Sample Manifest Tier
  pack {
    name          = "manifest-3"
    type          = "manifest"
    install_order = 0
    manifest {
      name    = "test-manifest-3"
      content = <<-EOT
          apiVersion: apps/v1
          kind: Deployment
          metadata:
            name: nginx-deployment
            labels:
              app: nginx
          spec:
            replicas: 3
            selector:
              matchLabels:
                app: nginx
            template:
              metadata:
                labels:
                  app: nginx
              spec:
                containers:
                  - name: nginx
                    image: nginx:1.14.2
                    ports:
                      - containerPort: 80
      EOT
    }
  }
  # Sample Operator-Instance Tier's
  pack {
    name            = "minio-operator-stage"
    type            = data.spectrocloud_pack_simple.minio_pack.type
    source_app_tier = data.spectrocloud_pack_simple.minio_pack.id
    properties = {
      "minioUsername"     = "miniostaging"
      "minioUserPassword" = base64encode("test123!wewe!")
      "volumeSize"        = "10"
    }
  }
  pack {
    name            = "mysql-3-stage"
    type            = data.spectrocloud_pack_simple.mysql_pack.type
    source_app_tier = data.spectrocloud_pack_simple.mysql_pack.id
    properties = {
      "dbRootPassword" = base64encode("test123!wewe!")
      "dbVolumeSize"   = "20"
      "version"        = "5.7"
    }
  }
  pack {
    name            = "redis-4-stage"
    type            = data.spectrocloud_pack_simple.redis_pack.type
    source_app_tier = data.spectrocloud_pack_simple.redis_pack.id
    properties = {
      "databaseName"       = "redsitstaging"
      "databaseVolumeSize" = "10"
    }
  }
}
