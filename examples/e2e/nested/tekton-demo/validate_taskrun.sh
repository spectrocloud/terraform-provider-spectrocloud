#!/bin/bash

# Verify latest TaskRun
taskrun=$(kubectl -n tekton-trigger-demo get taskrun | tail -n 1 | awk '{print $1;}')
uid=$(kubectl -n tekton-trigger-demo get taskrun "$taskrun" -o jsonpath='{.metadata.uid}')
kubectl -n tekton-trigger-demo get taskrun "$taskrun" -o jsonpath="{.metadata.annotations.chains\.tekton\.dev/payload-taskrun-$uid}" | base64 -d > payload
signature=$(kubectl -n tekton-trigger-demo get taskrun "$taskrun" -o jsonpath="{.metadata.annotations.chains\.tekton\.dev/signature-taskrun-$uid}")
cosign verify-blob --key k8s://tekton-chains/signing-secrets --signature "$signature" ./payload
rm -f payload