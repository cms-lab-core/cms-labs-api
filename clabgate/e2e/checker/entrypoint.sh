#!/bin/sh
set -eu

echo "checker completed for ${ATTEMPT_ID:-unknown}"

payload=$(printf '{"max_score":10,"current_score":9,"result_display":"local smoke: 9/10","report":"kind end-to-end smoke passed","tasks":[{"title":"Session namespace","description":"checker received the session context","logs":[{"node":"smoke","namespace":"%s","message":"context is available"}],"complete":true}]}' "${SESSION_NAMESPACE:-unknown}")
digest=$(printf '%s' "$payload" | sha256sum | cut -d ' ' -f 1)
encoded=$(printf '%s' "$payload" | base64 | tr -d '\n')

# One record suffices for this small E2E result. Larger ones use 2048-char chunks.
printf 'CMS_LABS_CHECKER_RESULT_V1 %s %s\n' "$digest" "$encoded"
