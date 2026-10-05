# GitHub CI/CD

Актуальный pipeline находится в `.github/workflows`. Старый `.gitlab-ci.yml` сохранён только на переходный период и не является источником истины.

## Проверки pull request

- `CI / Go lint` — `golangci-lint`, форматирование и синхронизация Go workspace;
- `CI / Go test` — MySQL 9.1, миграции backend и PNETLab addon, тесты и coverage artifact;
- `CI / Frontend lint, types and build` — ESLint, TypeScript, Prettier и production build;
- `CI / Helm validation` — lint всех конфигураций `pre` и `master`;
- `Kubernetes E2E / Session lifecycle on kind` — одноразовый Kubernetes в Docker, две реплики Clabgate, Clabernetes, Jupyter workspace, checker и обратная отправка оценки в mock CMS;
- `CodeQL` — анализ Go и JavaScript/TypeScript.

E2E не использует production kubeconfig: кластер `kind` создаётся внутри GitHub-hosted runner и удаляется вместе с runner после job. При ошибке workflow печатает Kubernetes events, describe и логи контейнеров.

## Образы

Workflow `Container images` проверяет Docker build в pull request и после push публикует в GHCR:

```text
ghcr.io/cms-lab-core/cms-labs-api/backend
ghcr.io/cms-lab-core/cms-labs-api/clabgate
ghcr.io/cms-lab-core/cms-labs-api/frontend
```

Для каждой ветки создаётся одноимённый tag (`main`, `pre`, `stage`), для каждого commit — `sha-<short-sha>`, для Git tag `v1.2.3` — `v1.2.3`. Стабильный mutable alias `latest` обновляется только при push Git tag `v*`; push ветки его не изменяет. Вместе с опубликованными образами BuildKit генерирует provenance и SBOM.

Workflow использует стандартный `GITHUB_TOKEN`; отдельный пароль GHCR не нужен. Если образы должны скачиваться Kubernetes без `imagePullSecret`, GitHub Packages нужно сделать public.

## Helm OCI chart

При push Git tag `vX.Y.Z` workflow `Helm OCI chart` проверяет, упаковывает и публикует `k8s/base-chart` с той же SemVer-версией:

```bash
helm pull oci://ghcr.io/cms-lab-core/cms-labs-api/charts/universal-chart --version X.Y.Z
helm upgrade --install core-backend \
  oci://ghcr.io/cms-lab-core/cms-labs-api/charts/universal-chart \
  --version X.Y.Z \
  --values k8s/values/_common/_common.yaml \
  --values k8s/values/_common/core-backend-values.yaml \
  --values k8s/values/master/core-backend-values.yaml
```

Для повторной публикации без Git tag workflow можно запустить вручную с SemVer в `chart_version`.

## PNETLab Debian package

`PNETLab addon package` запускается вручную для существующего Git tag — достаточно указать `release_tag`. `pnetlab-inject.zip` хранится в репозитории (`pnetlabaddon/pnetlabaddon/etc/pnetlabaddon/pnetlab-inject.zip`) и попадает внутрь Debian-пакета. Workflow собирает `pnetlabaddon.deb` и прикладывает к GitHub Release оба ассета: `pnetlabaddon.deb` и `pnetlab-inject.zip`.

Установленный addon обновляется из GitHub Releases:

```bash
./updater.sh                         # latest release
./updater.sh --package-version v1.2.3
```

## Обновления зависимостей

Dependabot раз в неделю проверяет GitHub Actions, все Go modules, frontend npm/yarn зависимости и Docker base images. Версии Actions в workflow зафиксированы точными release tags, а обновления приходят отдельными pull request.

Инструкция первичной настройки репозитория, branch protection, GHCR и environments: [GITHUB.md](GITHUB.md).
