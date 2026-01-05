---
page_title: "spectrocloud_application_profile Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  Provisions an Application Profile. App Profiles are templates created with preconfigured services. You can create as many profiles as required, with multiple tiers serving different functionalities per use case.
---

# spectrocloud_application_profile (Resource)

  Provisions an Application Profile. App Profiles are templates created with preconfigured services. You can create as many profiles as required, with multiple tiers serving different functionalities per use case.

## Example Usage


### App Profile with Manifest

```hcl
data "spectrocloud_pack" "byom" {
  name = "spectro-byo-manifest"
  version  = "1.0.0"
}

data "spectrocloud_pack" "csi" {
  name = "csi-gcp"
  version  = "1.0.0"
}

data "spectrocloud_pack" "cni" {
  name    = "cni-calico"
  version = "3.16.0"
}

data "spectrocloud_pack" "k8s" {
  name    = "kubernetes"
  version = "1.18.14"
}

data "spectrocloud_pack" "ubuntu" {
  name = "ubuntu-gcp"
  version  = "1.0.0"
}

resource "spectrocloud_application_profile" "profile" {
  name        = "gcp-picard-2"
  description = "basic cp"
  version     = "1.0.0"
  context     = "tenant"
  cloud       = "all"

  pack {
    name            = "manifest-1"
    tag             = "1.0.0"
    type            = "manifest"
    source_app_tier = "spectro-manifest-pack"

    values = <<-EOT
      manifests:
        byo-manifest:
          contents: |
            apiVersion: v1
            kind: Namespace
            metadata:
              labels:
                app: wordpress
                app3: wordpress3
              name: wordpress
    EOT
  }
}
   
```

### App Profile with Container

```hcl

variable "single-container-image" {
    type        = string
    description = "The name of the container image to use for the virtual cluster in a single scenario"
    default     = "ghcr.io/spectrocloud/hello-universe:1.0.8"
}

data "spectrocloud_cluster_group" "beehive" {
  name    = "beehive"
  context = "system"
}

data "spectrocloud_registry" "container_registry" {
  name = "Public Repo"
}

data "spectrocloud_pack_simple" "container_pack" {
  type         = "container"
  name         = "container"
  version      = "1.0.0"
  registry_uid = data.spectrocloud_registry.container_registry.id
}

resource "spectrocloud_application_profile" "hello-universe-ui" {
  name        = "hello-universe"
  description = "Hello Universe as a single UI instance"
  pack {
    name = "hello-universe-ui"
    type = data.spectrocloud_pack_simple.container_pack.type
    registry_uid = data.spectrocloud_registry.container_registry.id
    source_app_tier = data.spectrocloud_pack_simple.container_pack.id
    values = <<-EOT
        containerService:
            serviceName: "hello-universe-ui-test"
            registryUrl: ""
            image: ${var.single-container-image}
            access: public
            ports:
              - "8080"
            serviceType: LoadBalancer
    EOT
  }
  tags = ["scenario-1"]
}


```


###  App Profile with Helm Chart

```hcl
data "spectrocloud_registry_helm" "helm" {
  name = "Public Spectro Helm Repo"
}

data "spectrocloud_pack" "marvel-app" {
  name         = "marvel-app"
  registry_uid = data.spectrocloud_registry_helm.helm.id
  version      = "0.1.0"
}

resource "spectrocloud_application_profile" "my-app-profile" {
  name        = "my-app-profile"
  description = "A profile for a simple application"
  context     = "project"
  pack {
    name            = data.spectrocloud_pack.marvel-app.name
    type            = "helm"
    registry_uid    = data.spectrocloud_registry_helm.helm.id
    source_app_tier = data.spectrocloud_pack.marvel-app.id
  }
  tags = ["name:marvel-app", "terraform_managed:true", "env:dev"]
}
```

###  App Profile with All Type of Tiers
```hcl

data "spectrocloud_registry" "common_registry" {
  name = "Public Repo"
}

data "spectrocloud_registry" "container_registry" {
  name = "Public Repo"
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
            serviceName: "spectro-system-appdeployment-tiername-svc"
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

```

```
## Import

# terraform import spectrocloud_application_profile.app_profile_all_tiers "profile_uid_here"
# 
# Where:
# - profile_uid_here is the unique identifier of the application profile
#
# To import using import block:
# import {
#   to = spectrocloud_application_profile.app_profile_all_tiers
#   id = "profile_uid_here"
# }
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the application profile
- `pack` (Block Set, Min: 1) A list of packs to be applied to the application profile. (see [below for nested schema](#nestedblock--pack))

### Optional

- `cloud` (String) The cloud provider the profile is eligible for. Default value is `all`.
- `context` (String) Context of the profile. Allowed values are `project`, `cluster`, or `namespace`. Default value is `project`.If  the `project` context is specified, the project name will sourced from the provider configuration parameter [`project_name`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs#schema).
- `description` (String) Description of the profile.
- `tags` (Set of String) A list of tags to be applied to the application profile. Tags must be in the form of `key:value`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `version` (String) Version of the profile. Default value is 1.0.0.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- `name` (String) The name of the specified pack.

Optional:

- `install_order` (Number) The installation priority order of the app profile. The order of priority goes from lowest number to highest number. For example, a value of `-3` would be installed before an app profile with a higher number value. No upper and lower limits exist, and you may specify positive and negative integers. The default value is `0`.
- `manifest` (Block List) The manifest of the pack. (see [below for nested schema](#nestedblock--pack--manifest))
- `properties` (Map of String) The various properties required by different database tiers eg: `databaseName` and `databaseVolumeSize` size for Redis etc.
- `registry_name` (String) The name of the registry to be used for the pack. This can be used instead of `registry_uid` for better readability. Either `registry_uid` or `registry_name` can be specified, but not both.
- `registry_uid` (String) The unique id of the registry to be used for the pack. Either `registry_uid` or `registry_name` can be specified, but not both.
- `source_app_tier` (String) The unique id of the pack to be used as the source for the pack.
- `tag` (String) The identifier or version to label the pack.
- `type` (String) The type of Pack. Allowed values are `container`, `helm`, `manifest`, or `operator-instance`.
- `uid` (String) The unique id of the pack. This is a computed field and is not required to be set.
- `values` (String) The values to be used for the pack. This is a stringified JSON object.

<a id="nestedblock--pack--manifest"></a>
### Nested Schema for `pack.manifest`

Required:

- `content` (String) The content of the manifest.
- `name` (String) The name of the manifest.

Read-Only:

- `uid` (String)



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)