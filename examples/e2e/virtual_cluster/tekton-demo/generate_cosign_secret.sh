#!/bin/bash

# Generate x509 keypair & cosign secret
COSIGN_PASSWORD="" cosign generate-key-pair
cosign_pub=$(base64 < cosign.pub)
cosign_key=$(base64 < cosign.key)

cat <<EOF > manifests/secret-cosign.yaml
apiVersion: v1
data:
  cosign.key: $cosign_key
  cosign.password: ""
  cosign.pub: $cosign_pub
immutable: true
kind: Secret
metadata:
  name: signing-secrets
  namespace: tekton-chains
type: Opaque
EOF

rm -f cosign.key cosign.pub