---
page_title: "spectrocloud_application_profile Resource - terraform-provider-spectrocloud"
subcategory: "App Profiles"
description: |-
  Provisions an Application Profile. App Profiles are templates created with preconfigured services. You can create as many profiles as required, with multiple tiers serving different functionalities per use case.
---

# spectrocloud_application_profile (Resource)

  Provisions an Application Profile. App Profiles are templates created with preconfigured services. You can create as many profiles as required, with multiple tiers serving different functionalities per use case.

## Example Usage

```terraform
data "spectrocloud_registry" "common_registry" {
  name = "Public Repo"
}

data "spectrocloud_registry" "container_registry" {
  name = "automation-pack-registry"
}

data "spectrocloud_registry" "db_registry" {
  name = "svtest"
}

data "spectrocloud_registry" "bitnami_registry" {
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
  registry_uid = data.spectrocloud_registry.bitnami_registry.id
}

# This profile demonstrates every pack tier type an application profile supports: a plain
# container image, a Helm chart, a raw Kubernetes manifest, and three operator-instance tiers
# (each backed by a Palette operator pack). Nothing on this resource is ForceNew - the whole
# profile, including its packs, updates in place.
#
#   name        - Required.
#   version     - Optional, default "1.0.0". Must be a valid (or coercible) semantic version.
#   context     - Optional, default "project". Allowed: "project", "tenant", "system".
#   tags        - Optional. Tags are conventionally "key:value" strings.
#   description - Optional.
#   cloud       - Optional, default "all". The cloud provider this profile is eligible for.
#
# Common pack fields not called out per tier below: name (Required, unique per profile), uid
# (Computed - don't set), registry_name (Optional, mutually exclusive with registry_uid), tag
# (Optional), manifest (Optional, one or more raw-manifest blocks with name/content).
resource "spectrocloud_application_profile" "app_profile_all_tiers" {
  name        = "profile-all-tiers-test"
  version     = "1.0.0"
  context     = "project"
  tags        = ["owner:sivaa", "managed-by:terraform"]
  description = "Application profile demonstrating container, Helm, manifest, and operator-instance tiers."
  cloud       = "all"

  # pack (container-tier, container image tier):
  #   type            - Optional, default "spectro"; set explicitly here to "container".
  #   registry_uid    - Optional, mutually exclusive with registry_name.
  #   source_app_tier - Optional. UID of the source pack this tier is based on.
  #   values          - Optional. Pack configuration values in YAML/JSON.
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

  # pack (kafka-tier, Helm chart tier):
  #   type            - Optional, default "spectro"; set explicitly here to "helm".
  #   registry_uid    - Optional, mutually exclusive with registry_name.
  #   source_app_tier - Optional. UID of the source pack this tier is based on.
  pack {
    name            = "kafka-tier"
    type            = data.spectrocloud_pack_simple.kafka_pack.type
    registry_uid    = data.spectrocloud_registry.bitnami_registry.id
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

  # pack (manifest-3, raw Kubernetes manifest tier):
  #   type          - Optional, default "spectro"; set explicitly here to "manifest".
  #   install_order - Optional, default 0. Lower values run first.
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

  # pack (minio-operator-stage, operator-instance tier):
  #   type            - Optional, default "spectro"; set explicitly here to "operator-instance".
  #   source_app_tier - Optional. UID of the source pack this tier is based on.
  #   properties      - Optional. Simple key-value pack inputs (as opposed to YAML `values`).
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

  # pack (mysql-3-stage, operator-instance tier):
  #   type            - Optional, default "spectro"; set explicitly here to "operator-instance".
  #   source_app_tier - Optional. UID of the source pack this tier is based on.
  #   properties      - Optional. Simple key-value pack inputs (as opposed to YAML `values`).
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

  # pack (redis-4-stage, operator-instance tier):
  #   type            - Optional, default "spectro"; set explicitly here to "operator-instance".
  #   source_app_tier - Optional. UID of the source pack this tier is based on.
  #   properties      - Optional. Simple key-value pack inputs (as opposed to YAML `values`).
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

# Import example:
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

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import)
to import the resource spectrocloud_application_profile by using its `id` with the Palette `context` separated by a colon. For example:

```terraform
import {
  to = spectrocloud_application_profile.app_profile_all_tiers
  id = "profile_uid_here" 
}
```

Using `terraform import`, import the application profile using the `profile_uid_here` or `profile_name_here` colon separated with `context`. For example:

```console
terraform import spectrocloud_application_profile.example profile_uid_here/profile_name_here
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Readable name for the application profile.
- `pack` (Block List, Min: 1) A list of packs to be applied to the application profile. (see [below for nested schema](#nestedblock--pack))

### Optional

- `cloud` (String) The cloud provider the profile is eligible for. Default value is `all`.
- `context` (String) Context of the profile. Allowed values are `project`, `tenant`, or `system`. Default value is `project`.If  the `project` context is specified, the project name will sourced from the provider configuration parameter [`project_name`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs#schema).
- `description` (String) Description of the profile.
- `tags` (Set of String) A list of tags to be applied to the application profile. Tags must be in the form of `key:value`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `version` (String) Version of the profile. Default value is 1.0.0. Must be a valid semantic version (e.g. `1.2.3`) or a short/coerced form (e.g. `1`, `1.2`, `v1.2.3`) - malformed versions (e.g. `1.2.3.beta`, `V1.2.3`, `chart-v1.2.3`) are rejected by the API on create and update.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--pack"></a>
### Nested Schema for `pack`

Required:

- `name` (String) The name of the specified pack.

Optional:

- `install_order` (Number) The installation priority order of the app profile. The order of priority goes from lowest number to highest number. For example, a value of `-3` would be installed before an app profile with a higher number value. No upper and lower limits exist, and you may specify positive and negative integers. The default value is `0`.
- `manifest` (Block List) The manifest of the pack. (see [below for nested schema](#nestedblock--pack--manifest))
- `properties` (Map of String) Map of property name to string value required by the pack tier (for example, `databaseName` or `databaseVolumeSize`).
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