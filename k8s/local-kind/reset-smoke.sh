#!/bin/sh
set -eu

context=kind-cms-labs-local
system_namespace=cms-labs-system
session_namespace=lab-550e8400-e29b-41d4-a716-446655440000

# The namespace contains only disposable smoke-test state.
kubectl delete namespace "$session_namespace" \
  --context "$context" --ignore-not-found --wait
kubectl apply --context "$context" -f "$(dirname "$0")/smoke.yaml"
kubectl rollout restart deployment/clabgate deployment/front deployment/smoke-mock \
  --namespace "$system_namespace" --context "$context"
