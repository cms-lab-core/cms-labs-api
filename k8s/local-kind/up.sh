#!/bin/sh
set -eu

repo_root=$(CDPATH='' cd -- "$(dirname "$0")/../.." && pwd)
cluster_name=cms-labs-local
context="kind-$cluster_name"
state_dir=${CMS_LABS_DEV_STATE_DIR:-$repo_root/.local}
kubeconfig="$state_dir/kubeconfig"
node_image=${KIND_NODE_IMAGE:-kindest/node:v1.33.12@sha256:3f5c8443c620245e4d355cfe09e96a91ead32ceaa569d3f1ca9edf0cb2fe2ff4}
chart=${CLABERNETES_CHART:-oci://ghcr.io/cms-lab-core/cms-labs-clabernetes/clabernetes}
chart_version=${CLABERNETES_CHART_VERSION:-0.0.0}
dev_mode=false

usage() {
  cat <<EOF
Usage: $0 [--dev|--local]

  --dev, --local  Install the in-kind development gateway at http://127.0.0.1:18080.
                  It proxies host Vite/backend/clabgate and namespace-local
                  JupyterLab/ttyd Services through one browser origin.
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dev|--local)
      dev_mode=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Required command is not installed: $1" >&2
    exit 1
  fi
}

require_command docker
require_command kind
require_command kubectl
require_command helm

if ! docker info >/dev/null 2>&1; then
  echo "Docker daemon is not available" >&2
  exit 1
fi

if kind get clusters 2>/dev/null | grep -Fx "$cluster_name" >/dev/null 2>&1; then
  echo "Reusing kind cluster $cluster_name"
else
  kind create cluster \
    --config "$repo_root/k8s/local-kind/kind.yaml" \
    --image "$node_image"
fi

mkdir -p "$state_dir"
kind get kubeconfig --name "$cluster_name" > "$kubeconfig"

helm upgrade --install clabernetes "$chart" \
  --version "$chart_version" \
  --namespace c9s \
  --create-namespace \
  --kubeconfig "$kubeconfig" \
  --kube-context "$context"

kubectl apply --kubeconfig "$kubeconfig" --context "$context" -f "$repo_root/k8s/local-kind/dev.yaml"
if [ "$dev_mode" = true ]; then
  host_gateway=$(docker exec "${cluster_name}-control-plane" getent hosts host.docker.internal 2>/dev/null | awk 'NR == 1 { print $1 }')
  if [ -z "$host_gateway" ]; then
    host_gateway=$(docker inspect "${cluster_name}-control-plane" --format '{{range .NetworkSettings.Networks}}{{.Gateway}}{{end}}')
  fi
  if [ -z "$host_gateway" ]; then
    echo "Cannot determine the development host gateway address" >&2
    exit 1
  fi
  gateway_config="$state_dir/dev-gateway-config.yaml"
  kubectl create configmap cms-labs-dev-gateway \
    --namespace cms-labs-system \
    --from-file="default.conf.template=$repo_root/nextui-dashboard/nginx/templates/default.conf.template" \
    --from-file="frontend-location.conf=$repo_root/k8s/local-kind/dev-frontend-location.conf" \
    --dry-run=client \
    --output=yaml \
    --kubeconfig "$kubeconfig" \
    --context "$context" > "$gateway_config"
  kubectl apply \
    --kubeconfig "$kubeconfig" \
    --context "$context" \
    -f "$gateway_config"
  kubectl apply \
    --kubeconfig "$kubeconfig" \
    --context "$context" \
    -f "$repo_root/k8s/local-kind/dev-gateway.yaml"
  kubectl patch deployment/cms-labs-dev-gateway \
    --namespace cms-labs-system \
    --kubeconfig "$kubeconfig" \
    --context "$context" \
    --type merge \
    --patch "{\"spec\":{\"template\":{\"spec\":{\"hostAliases\":[{\"ip\":\"$host_gateway\",\"hostnames\":[\"host.docker.internal\"]}]}}}}"
  kubectl rollout restart deployment/cms-labs-dev-gateway \
    --namespace cms-labs-system \
    --kubeconfig "$kubeconfig" \
    --context "$context"
  kubectl rollout status deployment/cms-labs-dev-gateway \
    --namespace cms-labs-system \
    --kubeconfig "$kubeconfig" \
    --context "$context" \
    --timeout=2m
fi
kubectl wait --for=condition=Available deployment \
  --all \
  --namespace c9s \
  --kubeconfig "$kubeconfig" \
  --context "$context" \
  --timeout=5m

echo "Kind cluster $cluster_name is ready"
echo "Development kubeconfig: $kubeconfig"
if [ "$dev_mode" = true ]; then
  echo "Development gateway: http://127.0.0.1:18080"
  echo "Start in separate terminals: make dev-backend; make dev-clabgate; make dev-front"
fi
