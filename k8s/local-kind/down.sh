#!/bin/sh
set -eu

cluster_name=cms-labs-local

if ! command -v kind >/dev/null 2>&1; then
  echo "Required command is not installed: kind" >&2
  exit 1
fi

if kind get clusters 2>/dev/null | grep -Fx "$cluster_name" >/dev/null 2>&1; then
  kind delete cluster --name "$cluster_name"
else
  echo "Kind cluster $cluster_name does not exist"
fi
