#!/usr/bin/env bash
set -euo pipefail

package_version=latest
package_name=pnetlabaddon.deb
repository=cms-lab-core/cms-labs-api
api_url=https://api.github.com

while [[ $# -gt 0 ]]; do
  case "$1" in
    --package-version)
      package_version="${2:?--package-version requires a value}"
      shift 2
      ;;
    --repository)
      repository="${2:?--repository requires owner/repository}"
      shift 2
      ;;
    *)
      echo "Неизвестный параметр: $1" >&2
      exit 1
      ;;
  esac
done

headers=(
  --header "X-GitHub-Api-Version: 2022-11-28"
)
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  headers+=(--header "Authorization: Bearer $GITHUB_TOKEN")
fi

if [[ "$package_version" == latest ]]; then
  release_endpoint="$api_url/repos/$repository/releases/latest"
else
  encoded_tag=$(jq -rn --arg value "$package_version" '$value|@uri')
  release_endpoint="$api_url/repos/$repository/releases/tags/$encoded_tag"
fi

release=$(curl --fail --silent --show-error --location \
  "${headers[@]}" --header "Accept: application/vnd.github+json" "$release_endpoint")
asset_url=$(jq -er --arg name "$package_name" '.assets[] | select(.name == $name) | .url' <<<"$release" | head -n 1)
output="pnetlabaddon-${package_version}.deb"

curl --fail --silent --show-error --location \
  "${headers[@]}" \
  --header "Accept: application/octet-stream" \
  --output "$output" \
  "$asset_url"

echo "Файл успешно скачан: $output"
sudo dpkg -i "$output"
