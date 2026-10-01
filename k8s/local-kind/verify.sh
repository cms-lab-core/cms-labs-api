#!/bin/sh
set -eu

context=kind-cms-labs-local
namespace=lab-00000000-0000-0000-0000-000000000000
session_id=00000000-0000-0000-0000-000000000000
base_url=http://127.0.0.1:18080
repo_root=$(CDPATH='' cd -- "$(dirname "$0")/../.." && pwd)
cookie_jar=$(mktemp "${TMPDIR:-/tmp}/cms-labs-cookie.XXXXXX")
terminal_headers=$(mktemp "${TMPDIR:-/tmp}/cms-labs-terminal-headers.XXXXXX")
terminal_body=$(mktemp "${TMPDIR:-/tmp}/cms-labs-terminal-body.XXXXXX")
trap 'rm -f "$cookie_jar" "$terminal_headers" "$terminal_body"' EXIT HUP INT TERM

fail() {
  title=$1
  message=$2
  printf '::error title=%s::%s\n' "$title" "$message" >&2
  exit 1
}

assert_equal() {
  actual=$1
  expected=$2
  title=$3
  if test "$actual" != "$expected"; then
    fail "$title" "expected $expected, got $actual"
  fi
}

kubectl wait --for=condition=Available deployment/clabgate deployment/front deployment/smoke-mock \
  -n cms-labs-system --context "$context" --timeout=90s >/dev/null
kubectl wait --for=create deployment/jupyter \
  -n "$namespace" --context "$context" --timeout=90s >/dev/null
kubectl wait --for=condition=Available deployment/jupyter \
  -n "$namespace" --context "$context" --timeout=90s >/dev/null
kubectl wait --for=create topology/smoke \
  -n "$namespace" --context "$context" --timeout=90s >/dev/null
kubectl wait --for=jsonpath='{.status.topologyReady}'=true topology/smoke \
  -n "$namespace" --context "$context" --timeout=90s >/dev/null
kubectl get pvc/jupyter service/jupyter -n "$namespace" --context "$context" >/dev/null

desired_replicas=$(kubectl get deployment/clabgate -n cms-labs-system \
  --context "$context" -o jsonpath='{.spec.replicas}')
available_replicas=$(kubectl get deployment/clabgate -n cms-labs-system \
  --context "$context" -o jsonpath='{.status.availableReplicas}')
assert_equal "$desired_replicas" 2 "Unexpected Clabgate replica count"
assert_equal "$available_replicas" 2 "Clabgate replicas are not available"

attempt=0
holder=
while test -z "$holder"; do
  holder=$(kubectl get lease clabgate-session-reconciler -n cms-labs-system \
    --context "$context" -o jsonpath='{.spec.holderIdentity}' 2>/dev/null || true)
  attempt=$((attempt + 1))
  if test "$attempt" -ge 30; then
    fail "Clabgate leader was not elected" "Lease has no holder after 30 seconds"
  fi
  test -n "$holder" || sleep 1
done

attempt=0
while :; do
  state=$(curl --noproxy '*' --fail-with-body --silent --show-error "$base_url/api/state")
  attempt_status=$(printf '%s' "$state" | jq -r '.status')
  if test "$attempt_status" = active; then
    break
  fi
  attempt=$((attempt + 1))
  if test "$attempt" -ge 30; then
    fail "CMS attempt did not become active" "last status: $attempt_status"
  fi
  sleep 1
done

unauthorized=$(curl --noproxy '*' --silent --show-error --output /dev/null --write-out '%{http_code}' \
  "$base_url/clabgate/workspace/$session_id/")
assert_equal "$unauthorized" 401 "Unauthenticated workspace request was not rejected"

token=$(cd "$repo_root/clabgate" && \
  GOCACHE="${TMPDIR:-/tmp}/cms-labs-go-cache" go run ./e2e/token -env-file ../backend/.env.test)
authorization="Authorization: Bearer $token"
rpc_url="$base_url/clabgate/api/v1/rpc"

previous_check_id=$(printf '%s' "$state" | jq -r '.result.check_id // ""')
check_response=$(curl --noproxy '*' --fail-with-body --silent --show-error \
  -H "$authorization" -H 'Content-Type: application/json' \
  --data "{\"jsonrpc\":\"2.0\",\"id\":\"smoke-check\",\"method\":\"session.check\",\"params\":{\"session_id\":\"$session_id\"}}" \
  "$rpc_url")
job_name=$(printf '%s' "$check_response" | jq -er '.result.job_name')
kubectl wait --for=condition=complete "job/$job_name" -n "$namespace" \
  --context "$context" --timeout=90s >/dev/null

attempt=0
while :; do
  state=$(curl --noproxy '*' --fail-with-body --silent --show-error "$base_url/api/state")
  check_id=$(printf '%s' "$state" | jq -r '.result.check_id // ""')
  if test -n "$check_id" && test "$check_id" != "$previous_check_id"; then
    break
  fi
  attempt=$((attempt + 1))
  if test "$attempt" -ge 30; then
    fail "Checker result was not delivered" "CMS check_id did not change within 30 seconds"
  fi
  sleep 1
done

if ! printf '%s' "$state" | jq -e '
  .result.current_score == 9 and
  (.result.report | length > 0) and
  (.result.logs | length > 0) and
  .result.tasks[0].complete == true and
  .result.tasks[0].logs[0].message == "context is available"
' >/dev/null; then
  compact_state=$(printf '%s' "$state" | jq -c '{status, result}')
  fail "Invalid checker result" "expected structured checker result, got $compact_state"
fi

open_response=$(curl --noproxy '*' --fail-with-body --silent --show-error \
  -H "$authorization" -H 'Content-Type: application/json' \
  --data "{\"jsonrpc\":\"2.0\",\"id\":\"smoke-open\",\"method\":\"session.open\",\"params\":{\"session_id\":\"$session_id\"}}" \
  "$rpc_url")
open_url=$(printf '%s' "$open_response" | jq -er '.result.url')
exchange_status=$(curl --noproxy '*' --silent --show-error --cookie-jar "$cookie_jar" \
  --output /dev/null --write-out '%{http_code}' "$base_url$open_url")
assert_equal "$exchange_status" 303 "Workspace grant exchange did not redirect"
workspace_cookie=$(awk '$6 == "clabgate_workspace" {print $7}' "$cookie_jar")
if test -z "$workspace_cookie"; then
  fail "Workspace cookie is missing" "grant exchange did not set clabgate_workspace"
fi
workspace=$(curl --noproxy '*' --fail-with-body --silent --show-error \
  -H "Cookie: clabgate_workspace=$workspace_cookie" \
  "$base_url/clabgate/workspace/$session_id/")
if ! printf '%s' "$workspace" | grep -q 'workspace ready'; then
  fail "Workspace response is invalid" "expected workspace ready marker"
fi

topology_response=$(curl --noproxy '*' --fail-with-body --silent --show-error \
  -H "$authorization" -H 'Content-Type: application/json' \
  --data "{\"jsonrpc\":\"2.0\",\"id\":\"smoke-topology\",\"method\":\"topology.get\",\"params\":{\"session_id\":\"$session_id\"}}" \
  "$rpc_url")
terminal_grant=$(printf '%s' "$topology_response" | jq -er '.result.ttyd[0].url')
terminal_exchange_status=$(curl --noproxy '*' --silent --show-error \
  --dump-header "$terminal_headers" --output /dev/null --write-out '%{http_code}' \
  "$base_url$terminal_grant")
assert_equal "$terminal_exchange_status" 303 "Terminal grant exchange did not redirect"
terminal_location=$(awk 'tolower($1) == "location:" {sub(/\r$/, "", $2); print $2}' "$terminal_headers")
if test -z "$terminal_location"; then
  fail "Terminal redirect is missing" "grant exchange returned no Location header"
fi
curl --noproxy '*' --fail-with-body --silent --show-error \
  -H "Cookie: clabgate_workspace=$workspace_cookie" \
  --output "$terminal_body" "$base_url$terminal_location"
if ! grep -qi 'ttyd' "$terminal_body"; then
  fail "Terminal proxy response is invalid" "expected ttyd browser client"
fi

printf 'smoke passed: leader=%s job=%s check_id=%s score=%s\n' \
  "$holder" "$job_name" "$check_id" "$(printf '%s' "$state" | jq -r '.result.result_display')"
