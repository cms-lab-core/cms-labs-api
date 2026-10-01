# Перенос в github.com/cms-lab-core/cms-labs-api

Код подготовлен к GitHub: Go module paths, badge/link targets, CI, kind E2E, CodeQL, Dependabot и GHCR больше не зависят от GitLab.

## 1. Создать и отправить репозиторий

Создайте пустой repository `cms-lab-core/cms-labs-api` без автоматически добавленного README. Локальный clone уже настроен на HTTPS remote и ветку `main`; для первой отправки достаточно:

```bash
git push -u origin main
git push origin --tags
```

## 2. Настроить Actions и защиту ветки

В Settings → Actions → General разрешите Actions для repository. Затем создайте ruleset для default branch:

- запрет direct push;
- минимум один approve;
- dismiss stale approvals;
- require conversation resolution;
- require branch to be up to date;
- обязательные checks: `Pre-commit hooks`, `Go lint`, `Go test`, `Frontend lint, types and build`, `Helm validation`, `Session lifecycle on kind`, оба CodeQL checks и три `Build …` image checks.

AI review можно подключить позже отдельным GitHub App/check и добавить в тот же ruleset.

## 3. Настроить GHCR

Первый push в `main`, `pre` или `stage` создаст три container packages. Git tag `vX.Y.Z` дополнительно опубликует OCI Helm chart `ghcr.io/cms-lab-core/cms-labs-api/charts/universal-chart:X.Y.Z`. В каждом package откройте Package settings и включите `Inherit access from source repository`. Для кластеров без GHCR credentials выставьте Public visibility; для private packages создайте Kubernetes `imagePullSecret` и выполните `helm registry login` перед pull private chart.

## 4. Проверить миграцию

1. Откройте pull request и дождитесь всех CI и kind E2E checks.
2. Проверьте появление трёх образов и tags `sha-*` в Packages.
3. Создайте Git tag `vX.Y.Z` и проверьте публикацию OCI Helm chart.

## Совместимость источников заданий

`CMS_TASK_URL` принимает полный URL проекта GitLab или GitHub. Для private repository задайте `TASK_REPOSITORY_TOKEN`; старый `GITLAB_TOKEN` остаётся fallback на время миграции. Токен используется только для API выбранного provider и не попадает в namespace лаборатории.

## PNETLab addon

Сборка `.deb` включает `pnetlab-inject.zip`, который хранится прямо в репозитории по пути `pnetlabaddon/pnetlabaddon/etc/pnetlabaddon/pnetlab-inject.zip` (закоммичен, публичный). Запустите workflow `PNETLab addon package`, укажите существующий `release_tag` — workflow соберёт Debian-пакет и приложит к GitHub Release оба ассета: `pnetlabaddon.deb` и `pnetlab-inject.zip`. Никаких секретов и внешних хранилищ не требуется.

## Что осталось legacy

- `.gitlab-ci.yml` и каталог старых GitLab job сохранены для переходного периода, но GitHub Actions их не вызывает.
- Перед первым public push проверьте историю секрет-сканером. В тестовых fixtures есть RSA material для локальных JWT; если GitHub push protection сочтёт его секретом, перенесите fixture на runtime generation и очистите Git history отдельной согласованной операцией.
