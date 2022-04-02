data "spectrocloud_pack" "csi" {
  name    = var.csi_name
  version = var.csi_ver
}

data "spectrocloud_pack" "cni" {
  name    = var.cni_name
  version = var.cni_ver
}

data "spectrocloud_pack" "k8s" {
  name    = var.k8s_name
  version = var.k8s_ver
}

data "spectrocloud_pack" "os" {
  name    = var.os_name
  version = var.os_ver
}

data "spectrocloud_pack" "addon" {
  name    = var.addon_name
  version = var.addon_ver
}
resource "spectrocloud_cluster_profile" "profile" {
  name  = var.cp_name
  type  = var.type_name
  cloud = var.cloud_name

  pack {
    name = data.spectrocloud_pack.os.name
    tag  = var.os_ver
    uid  = data.spectrocloud_pack.os.id
  }
  pack {
    name   = data.spectrocloud_pack.k8s.name
    tag    = var.k8s_ver
    uid    = data.spectrocloud_pack.k8s.id
    values = <<-EOT
               manifests:
                      -manifest:
                          contents: |
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
                                    certSANs:
                                        - "cluster-{{ .spectro.system.cluster.uid }}.{{ .spectro.system.reverseproxy.server }}"
                                    extraArgs:
                                      # Note : secure-port flag is used during kubeadm init. Do not change this flag on a running cluster
                                      secure-port: "6443"
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
                                  scheduler:
                                    extraArgs:
                                      profiling: "false"
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
                                    # Sometimes api server takes a little longer to respond. Retry if applying the pod-security-policy manifest fails
                                    - 'export KUBECONFIG=/etc/kubernetes/admin.conf && [ -f "$KUBECONFIG" ] && { echo " ====> Applying PodSecurityPolicy" ; until $(kubectl apply -f /etc/kubernetes/hardening/privileged-psp.yaml > /dev/null ); do echo "Failed to apply PodSecurityPolicies, will retry in 5s" ; sleep 5 ; done ; } || echo "Skipping PodSecurityPolicy for worker nodes"'

                                # Client configuration to add OIDC based authentication flags in kubeconfig
                                #clientConfig:
                                  #oidc-issuer-url: "{{ .spectro.pack.kubernetes.kubeadmconfig.apiServer.extraArgs.oidc-issuer-url }}"
                                  #oidc-client-id: "{{ .spectro.pack.kubernetes.kubeadmconfig.apiServer.extraArgs.oidc-client-id }}"
                                  #oidc-client-secret: 1gsranjjmdgahm10j8r6m47ejokm9kafvcbhi3d48jlc3rfpprhv
                                  #oidc-extra-scope: profile,email
              EOT
  }

  pack {
    name = data.spectrocloud_pack.cni.name
    tag  = var.cni_ver
    uid  = data.spectrocloud_pack.cni.id
  }
  pack {
    name = data.spectrocloud_pack.csi.name
    tag  = var.csi_ver
    uid  = data.spectrocloud_pack.csi.id
  }
  pack {
    name   = data.spectrocloud_pack.addon.name
    tag    = var.addon_ver
    uid    = data.spectrocloud_pack.addon.id
    values = ""
  }
}
