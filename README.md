# CMS Labs API

[![CI](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/ci.yml/badge.svg)](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/ci.yml)
[![Kubernetes E2E](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/e2e.yml/badge.svg)](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/e2e.yml)
[![CodeQL](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/codeql.yml/badge.svg)](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/codeql.yml)
[![Container images](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/images.yml/badge.svg)](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/images.yml)
[![Helm OCI chart](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/helm-chart.yml/badge.svg)](https://github.com/cms-lab-core/cms-labs-api/actions/workflows/helm-chart.yml)

Монорепозиторий CMS Labs: Go backend, Clabgate, PNETLab addon, frontend и Kubernetes-конфигурация.
Основной адрес проекта: <https://github.com/cms-lab-core/cms-labs-api>.

Контейнеры публикуются в GitHub Container Registry:

- `ghcr.io/cms-lab-core/cms-labs-api/backend`;
- `ghcr.io/cms-lab-core/cms-labs-api/clabgate`;
- `ghcr.io/cms-lab-core/cms-labs-api/frontend`.

Стабильный контур использует образы релизного тега `vX.Y.Z`. Mutable-тег `latest` обновляется только при публикации Git tag `v*` и соответствует последнему стабильному релизу; push в ветку его не изменяет.

Проверки лабораторных изолированы в отдельном Go-репозитории
[`cms-labs-checker`](https://github.com/cms-lab-core/cms-labs-checker). Он
публикует единый образ `ghcr.io/cms-lab-core/cms-labs-checker`, внутри которого
каждая лабораторная имеет собственный пакет и unit-тесты.

Переиспользуемый Helm chart публикуется при Git tag `vX.Y.Z`:

```bash
helm pull oci://ghcr.io/cms-lab-core/cms-labs-api/charts/universal-chart --version X.Y.Z
```

## CI/CD

1. [Pipelines](CI-CD/README.md)
2. [Миграция и настройка GitHub](CI-CD/GITHUB.md)
3. [Docker files](CI-CD/docker.md)

## Монорепозиторий для модулей:

1. [CMS Labs Core](backend/README.md)
2. [CMS Labs Front](nextui-dashboard/README.md)
3. [PNET Lab Addon](pnetlabaddon/README.md)
4. [Go-Lang Shared](shared)
5. [Gen Model](gen/README.md)
6. [Clabgate: Kubernetes-сессии без JupyterHub, граф зависимостей и план миграции](clabgate/README.md)

## Документация методов Swagger

1. [CMS Labs Core](backend/docs/swagger.json)
2. [PNET Lab Addon](pnetlabaddon/docs/swagger.json)

## ⚡️ Быстрый старт

1. Для установки всех зависимостей GoLang выполните:

```bash
make dev
```

2. Для тестирования всего пакета достаточно вызывать:

```bash
# Тестирует код
# Пишет coverage
make test

# Форматеры Go и Frontend'а
make pre_commit

# Линтер и инъекции кода
make security
```

3. Для автоматической генерации документации везде:

```bash
make generate
```
