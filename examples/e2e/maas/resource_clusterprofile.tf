# If looking up a cluster profile instead of creating a new one
# data "spectrocloud_cluster_profile" "profile" {
#   # id = <uid>
#   name = var.cluster_cluster_profile_name
# }

# # Example of a Basic add-on profile
/*
		"profiles": [
			{
				"uid": "612fb02d1cbcb9ef7d4682ff",
				"packValues": [
					{
						"tag": "LTS__18.4.x",
						"name": "ubuntu-maas",
						"type": "spectro",
						"values": "# Spectro Golden images includes most of the hardening standards recommended by CIS benchmarking v1.5\n\n# Uncomment below section to\n# 1. Include custom files to be copied over to the nodes and/or\n# 2. Execute list of commands before or after kubeadm init/join is executed\n#\n#kubeadmconfig:\n#  preKubeadmCommands:\n#  - echo \"Executing pre kube admin config commands\"\n#  - update-ca-certificates\n#  - 'systemctl restart containerd; sleep 3'\n#  - 'while [ ! -S /var/run/containerd/containerd.sock ]; do echo \"Waiting for containerd...\"; sleep 1; done'\n#  postKubeadmCommands:\n#  - echo \"Executing post kube admin config commands\"\n#  files:\n#  - targetPath: /usr/local/share/ca-certificates/mycom.crt\n#    targetOwner: \"root:root\"\n#    targetPermissions: \"0644\"\n#    content: |\n#      -----BEGIN CERTIFICATE-----\n#      MIICyzCCAbOgAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl\n#      cm5ldGVzMB4XDTIwMDkyMjIzNDMyM1oXDTMwMDkyMDIzNDgyM1owFTETMBEGA1UE\n#      AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMdA\n#      nZYs1el/6f9PgV/aO9mzy7MvqaZoFnqO7Qi4LZfYzixLYmMUzi+h8/RLPFIoYLiz\n#      qiDn+P8c9I1uxB6UqGrBt7dkXfjrUZPs0JXEOX9U/6GFXL5C+n3AUlAxNCS5jobN\n#      fbLt7DH3WoT6tLcQefTta2K+9S7zJKcIgLmBlPNDijwcQsbenSwDSlSLkGz8v6N2\n#      7SEYNCV542lbYwn42kbcEq2pzzAaCqa5uEPsR9y+uzUiJpv5tDHUdjbFT8tme3vL\n#      9EdCPODkqtMJtCvz0hqd5SxkfeC2L+ypaiHIxbwbWe7GtliROvz9bClIeGY7gFBK\n#      jZqpLdbBVjo0NZBTJFUCAwEAAaMmMCQwDgYDVR0PAQH/BAQDAgKkMBIGA1UdEwEB\n#      /wQIMAYBAf8CAQAwDQYJKoZIhvcNAQELBQADggEBADIKoE0P+aVJGV9LWGLiOhki\n#      HFv/vPPAQ2MPk02rLjWzCaNrXD7aPPgT/1uDMYMHD36u8rYyf4qPtB8S5REWBM/Y\n#      g8uhnpa/tGsaqO8LOFj6zsInKrsXSbE6YMY6+A8qvv5lPWpJfrcCVEo2zOj7WGoJ\n#      ixi4B3fFNI+wih8/+p4xW+n3fvgqVYHJ3zo8aRLXbXwztp00lXurXUyR8EZxyR+6\n#      b+IDLmHPEGsY9KOZ9VLLPcPhx5FR9njFyXvDKmjUMJJgUpRkmsuU1mCFC+OHhj56\n#      IkLaSJf6z/p2a3YjTxvHNCqFMLbJ2FvJwYCRzsoT2wm2oulnUAMWPI10vdVM+Nc=\n#      -----END CERTIFICATE-----",
						"manifests": []
					},
					{
						"tag": "1.21.x",
						"name": "kubernetes",
						"type": "spectro",
						"values": "pack:\n  k8sHardening: True\n  #CIDR Range for Pods in cluster\n  # Note : This must not overlap with any of the host or service network\n  podCIDR: \"192.168.0.0/16\"\n  #CIDR notation IP range from which to assign service cluster IPs\n  # Note : This must not overlap with any IP ranges assigned to nodes for pods.\n  serviceClusterIpRange: \"10.96.0.0/12\"\n\n# KubeAdm customization for kubernetes hardening. Below config will be ignored if k8sHardening property above is disabled\nkubeadmconfig:\n  apiServer:\n    extraArgs:\n      # Note : secure-port flag is used during kubeadm init. Do not change this flag on a running cluster\n      secure-port: \"6443\"\n      anonymous-auth: \"true\"\n      insecure-port: \"0\"\n      profiling: \"false\"\n      disable-admission-plugins: \"AlwaysAdmit\"\n      default-not-ready-toleration-seconds: \"60\"\n      default-unreachable-toleration-seconds: \"60\"\n      enable-admission-plugins: \"AlwaysPullImages,NamespaceLifecycle,ServiceAccount,NodeRestriction,PodSecurityPolicy\"\n      audit-log-path: /var/log/apiserver/audit.log\n      audit-policy-file: /etc/kubernetes/audit-policy.yaml\n      audit-log-maxage: \"30\"\n      audit-log-maxbackup: \"10\"\n      audit-log-maxsize: \"100\"\n      authorization-mode: RBAC,Node\n      tls-cipher-suites: \"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256\"\n    extraVolumes:\n      - name: audit-log\n        hostPath: /var/log/apiserver\n        mountPath: /var/log/apiserver\n        pathType: DirectoryOrCreate\n      - name: audit-policy\n        hostPath: /etc/kubernetes/audit-policy.yaml\n        mountPath: /etc/kubernetes/audit-policy.yaml\n        readOnly: true\n        pathType: File\n  controllerManager:\n    extraArgs:\n      profiling: \"false\"\n      terminated-pod-gc-threshold: \"25\"\n      pod-eviction-timeout: \"1m0s\"\n      use-service-account-credentials: \"true\"\n      feature-gates: \"RotateKubeletServerCertificate=true\"\n  scheduler:\n    extraArgs:\n      profiling: \"false\"\n  kubeletExtraArgs:\n    read-only-port : \"0\"\n    event-qps: \"0\"\n    feature-gates: \"RotateKubeletServerCertificate=true\"\n    protect-kernel-defaults: \"true\"\n    tls-cipher-suites: \"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256\"\n  files:\n    - path: hardening/audit-policy.yaml\n      targetPath: /etc/kubernetes/audit-policy.yaml\n      targetOwner: \"root:root\"\n      targetPermissions: \"0600\"\n    - path: hardening/privileged-psp.yaml\n      targetPath: /etc/kubernetes/hardening/privileged-psp.yaml\n      targetOwner: \"root:root\"\n      targetPermissions: \"0600\"\n    - path: hardening/90-kubelet.conf\n      targetPath: /etc/sysctl.d/90-kubelet.conf\n      targetOwner: \"root:root\"\n      targetPermissions: \"0600\"\n  preKubeadmCommands:\n    # For enabling 'protect-kernel-defaults' flag to kubelet, kernel parameters changes are required\n    - 'echo \"====> Applying kernel parameters for Kubelet\"'\n    - 'sysctl -p /etc/sysctl.d/90-kubelet.conf'\n  postKubeadmCommands:\n    # Apply the privileged PodSecurityPolicy on the first master node ; Otherwise, CNI (and other) pods won't come up\n    - 'export KUBECONFIG=/etc/kubernetes/admin.conf'\n    # Sometimes api server takes a little longer to respond. Retry if applying the pod-security-policy manifest fails\n    - '[ -f \"$KUBECONFIG\" ] && { echo \" ====> Applying PodSecurityPolicy\" ; until $(kubectl apply -f /etc/kubernetes/hardening/privileged-psp.yaml > /dev/null ); do echo \"Failed to apply PodSecurityPolicies, will retry in 5s\" ; sleep 5 ; done ; } || echo \"Skipping PodSecurityPolicy for worker nodes\"'\n\n# Client configuration to add OIDC based authentication flags in kubeconfig\n#clientConfig:\n  #oidc-issuer-url: \"{{ .spectro.pack.kubernetes.kubeadmconfig.apiServer.extraArgs.oidc-issuer-url }}\"\n  #oidc-client-id: \"{{ .spectro.pack.kubernetes.kubeadmconfig.apiServer.extraArgs.oidc-client-id }}\"\n  #oidc-client-secret: 1gsranjjmdgahm10j8r6m47ejokm9kafvcbhi3d48jlc3rfpprhv\n  #oidc-extra-scope: profile,email",
						"manifests": []
					},
					{
						"tag": "3.19.x",
						"name": "cni-calico",
						"type": "spectro",
						"values": "manifests:\n  calico:\n\n    # IPAM type to use. Supported types are calico-ipam, host-local\n    ipamType: \"calico-ipam\"\n\n    # Should be one of CALICO_IPV4POOL_IPIP or CALICO_IPV4POOL_VXLAN\n    encapsulationType: \"CALICO_IPV4POOL_IPIP\"\n\n    # Should be one of Always, CrossSubnet, Never\n    encapsulationMode: \"Always\"",
						"manifests": []
					},
					{
						"tag": "1.0.x",
						"name": "csi-maas-volume",
						"type": "spectro",
						"values": "manifests:\n  maas:\n\n    #Toggle for Default class\n    isDefaultClass: \"true\"\n\n    #Supported binding modes are Immediate, WaitForFirstConsumer\n    volumeBindingMode: \"WaitForFirstConsumer\"",
						"manifests": []
					}
				]
			}
		],

*/


data "spectrocloud_pack" "csi" {
  name = "csi-maas-volume"
  # version  = "1.0.x"
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
  name = "ubuntu-maas"
  # version  = "1.0.x"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "maas-picard-3"
  description = "basic cp"
  cloud       = "maas"
  type        = "cluster"


  pack {
    name   = "csi-maas-volume"
    tag    = "1.0.x"
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }

  pack {
    name   = "cni-calico"
    tag    = "3.16.x"
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = "kubernetes"
    tag    = "1.18.x"
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = "ubuntu-maas"
    tag    = "LTS__18.4.x"
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }
}
