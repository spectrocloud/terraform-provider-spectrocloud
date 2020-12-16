resource "spectrocloud_cluster_profile" "cp-basic" {
  name        = "cp-basic"
  description = "basic cp"
  cloud       = "azure"

  pack {
    name = "spectro-byo-manifest"
    tag  = "1.0.x"
    uid  = "5faad584f244cfe0b98cf489"
    # layer  = ""
    values = <<-EOT
      manifests:
        byo-manifest:
          contents: |
            # Add manifests here
            apiVersion: v1
            kind: Namespace
            metadata:
              labels:
                app: wordpress
                app3: wordpress3
              name: wordpress
    EOT
  }

  pack {
    name = "csi-azure"
    tag  = "1.0.x"
    uid  = "5f7e5fc9b0e4543be6fc7d0f"
    values = <<-EOT
      manifests:

        azure_disk:

          # Azure storage account Sku tier. Default is empty
          storageaccounttype: "Standard_LRS"

          # Possible values are shared (default), dedicated, and managed
          kind: "managed"

          #Allowed reclaim policies are Delete, Retain
          reclaimPolicy: "Delete"

          #Toggle for Volume expansion
          allowVolumeExpansion: "true"

          #Toggle for Default class
          isDefaultClass: "true"

          #Supported binding modes are Immediate, WaitForFirstConsumer
          #Setting binding mode to WaitForFirstConsumer, so that the volumes gets created in the same AZ as that of the pods
          volumeBindingMode: "WaitForFirstConsumer"
    EOT
  }

  pack {
    name = "cni-calico-azure"
    tag  = "3.16.x"
    uid  = "5fd0ca727c411c71b55a359c"
    values = <<-EOT
      manifests:
        calico:

          # IPAM type to use. Supported types are calico-ipam, host-local
          ipamType: "host-local"

          # Uncomment property below when ipamType is set to calico-ipam
          # networkCIDR to use (should match the kubernetes podCIDR)
          #calicoNetworkCIDR: "192.168.0.0/16"

          # Should be one of CALICO_IPV4POOL_IPIP or CALICO_IPV4POOL_VXLAN
          encapsulationType: "CALICO_IPV4POOL_VXLAN"

          # Should be one of Always, CrossSubnet, Never
          encapsulationMode: "Always"
    EOT
  }

  pack {
    name = "kubernetes"
    tag  = "1.18.x"
    uid  = "5fd75aec3a52041fd14037ec"
    values = <<-EOT
      pack:
        k8sHardening: True
        #CIDR Range for Pods in cluster
        # Note : This must not overlap with any of the host or service network
        podCIDR: "192.168.0.0/16"
        #CIDR notation IP range from which to assign service cluster IPs
        # Note : This must not overlap with any IP ranges assigned to nodes for pods.
        serviceClusterIpRange: "10.96.0.0/12"

      # KubeAdm customization for kubernetes hardening. Below config will be ignored if k8sHardening property above is disabled
      kubeadmconfig:
        apiServer:
          extraArgs:
            anonymous-auth: "true"
            insecure-port: "0"
            profiling: "false"
            disable-admission-plugins: "AlwaysAdmit"
            default-not-ready-toleration-seconds: "60"
            default-unreachable-toleration-seconds: "60"
            enable-admission-plugins: "AlwaysPullImages,NamespaceLifecycle,ServiceAccount,NodeRestriction,PodSecurityPolicy"
            audit-log-path: /var/log/apiserver/audit.log
            audit-policy-file: /etc/kubernetes/audit-policy.yaml
            audit-log-maxage: "30"
            audit-log-maxbackup: "10"
            audit-log-maxsize: "100"
            authorization-mode: RBAC,Node
            tls-cipher-suites: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256"
          extraVolumes:
            - name: audit-log
              hostPath: /var/log/apiserver
              mountPath: /var/log/apiserver
              pathType: DirectoryOrCreate
            - name: audit-policy
              hostPath: /etc/kubernetes/audit-policy.yaml
              mountPath: /etc/kubernetes/audit-policy.yaml
              readOnly: true
              pathType: File
        controllerManager:
          extraArgs:
            profiling: "false"
            terminated-pod-gc-threshold: "25"
            pod-eviction-timeout: "1m0s"
            use-service-account-credentials: "true"
            feature-gates: "RotateKubeletServerCertificate=true"
            address: "0.0.0.0"
        scheduler:
          extraArgs:
            profiling: "false"
            address: "0.0.0.0"
        kubeletExtraArgs:
          read-only-port : "0"
          event-qps: "0"
          feature-gates: "RotateKubeletServerCertificate=true"
          protect-kernel-defaults: "true"
          tls-cipher-suites: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256"
        files:
          - path: hardening/audit-policy.yaml
            targetPath: /etc/kubernetes/audit-policy.yaml
            targetOwner: "root:root"
            targetPermissions: "0600"
          - path: hardening/privileged-psp.yaml
            targetPath: /etc/kubernetes/hardening/privileged-psp.yaml
            targetOwner: "root:root"
            targetPermissions: "0600"
          - path: hardening/90-kubelet.conf
            targetPath: /etc/sysctl.d/90-kubelet.conf
            targetOwner: "root:root"
            targetPermissions: "0600"
        preKubeadmCommands:
          # For enabling 'protect-kernel-defaults' flag to kubelet, kernel parameters changes are required
          - 'echo "====> Applying kernel parameters for Kubelet"'
          - 'sysctl -p /etc/sysctl.d/90-kubelet.conf'
        postKubeadmCommands:
          # Apply the privileged PodSecurityPolicy on the first master node ; Otherwise, CNI (and other) pods won't come up
          - 'export KUBECONFIG=/etc/kubernetes/admin.conf'
          # Sometimes api server takes a little longer to respond. Retry if applying the pod-security-policy manifest fails
          - '[ -f "$KUBECONFIG" ] && { echo " ====> Applying PodSecurityPolicy" ; until $(kubectl apply -f /etc/kubernetes/hardening/privileged-psp.yaml > /dev/null ); do echo "Failed to apply PodSecurityPolicies, will retry in 5s" ; sleep 5 ; done ; } || echo "Skipping PodSecurityPolicy for worker nodes"'
    EOT
  }

  pack {
    name = "ubuntu-azure"
    tag  = "LTS__18.4.x"
    uid  = "5f7e5fc9b0e4543ced7ed3b3"
    values = <<-EOT
      # Spectro Golden images includes most of the hardening standards recommended by CIS benchmarking v1.5

      # Uncomment below section to
      # 1. Include custom files to be copied over to the nodes and/or
      # 2. Execute list of commands before or after kubeadm init/join is executed
      #
      #kubeadmconfig:
      #  preKubeadmCommands:
      #  - echo "Executing pre kube admin config commands"
      #  - update-ca-certificates
      #  - 'systemctl restart containerd; sleep 3'
      #  - 'while [ ! -S /var/run/containerd/containerd.sock ]; do echo "Waiting for containerd..."; sleep 1; done'
      #  postKubeadmCommands:
      #  - echo "Executing post kube admin config commands"
      #  files:
      #  - targetPath: /usr/local/share/ca-certificates/mycom.crt
      #    targetOwner: "root:root"
      #    targetPermissions: "0644"
      #    content: |
      #      -----BEGIN CERTIFICATE-----
      #      MIICyzCCAbOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl
      #      cm5ldGVzMB4XDTIwMDkyMjIzNDMyM1oXDTMwMDkyMDIzNDgyM1owFTETMBEGA1UE
      #      AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMdA
      #      nZYs1el/6f9PgV/aO9mzy7MvqaZoFnqO7Qi4LZfYzixLYmMUzi+h8/RLPFIoYLiz
      #      qiDn+P8c9I1uxB6UqGrBt7dkXfjrUZPs0JXEOX9U/6GFXL5C+n3AUlAxNCS5jobN
      #      fbLt7DH3WoT6tLcQefTta2K+9S7zJKcIgLmBlPNDijwcQsbenSwDSlSLkGz8v6N2
      #      7SEYNCV542lbYwn42kbcEq2pzzAaCqa5uEPsR9y+uzUiJpv5tDHUdjbFT8tme3vL
      #      9EdCPODkqtMJtCvz0hqd5SxkfeC2L+ypaiHIxbwbWe7GtliROvz9bClIeGY7gFBK
      #      jZqpLdbBVjo0NZBTJFUCAwEAAaMmMCQwDgYDVR0PAQH/BAQDAgKkMBIGA1UdEwEB
      #      /wQIMAYBAf8CAQAwDQYJKoZIhvcNAQELBQADggEBADIKoE0P+aVJGV9LWGLiOhki
      #      HFv/vPPAQ2MPk02rLjWzCaNrXD7aPPgT/1uDMYMHD36u8rYyf4qPtB8S5REWBM/Y
      #      g8uhnpa/tGsaqO8LOFj6zsInKrsXSbE6YMY6+A8qvv5lPWpJfrcCVEo2zOj7WGoJ
      #      ixi4B3fFNI+wih8/+p4xW+n3fvgqVYHJ3zo8aRLXbXwztp00lXurXUyR8EZxyR+6
      #      b+IDLmHPEGsY9KOZ9VLLPcPhx5FR9njFyXvDKmjUMJJgUpRkmsuU1mCFC+OHhj56
      #      IkLaSJf6z/p2a3YjTxvHNCqFMLbJ2FvJwYCRzsoT2wm2oulnUAMWPI10vdVM+Nc=
      #      -----END CERTIFICATE-----
    EOT
  }

}
