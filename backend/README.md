# CMS Labs

- Core уровень для интеграции между LMS системой (Moodle) по протоколам LTI
- Предоставление SSO OPENID формата
- Управление серверами и кластерами
- Распределение пользователей между сегментами сервером
  (пока только PNETLab)

Шаблон [Create Go App CLI](https://github.com/create-go-app/cli)

<img src="https://img.shields.io/badge/Go-1.19+-00ADD8?style=for-the-badge&logo=go" alt="go version" />
&nbsp;<a href="https://goreportcard.com/report/github.com/cms-lab-core/cms-labs-api/backend" target="_blank"><img src="https://img.shields.io/badge/Go_report-A+-success?style=for-the-badge&logo=none" alt="go report" /></a>
&nbsp;<img src="https://img.shields.io/badge/license-Apache_2.0-red?style=for-the-badge&logo=none" alt="license" />

## ⚡️ Быстрый старт

1. Использовать для быстрого создания моделей:

```bash
MODEL='book' make crud
```

2. Для миграции используй:

```bash
make migrate.up
```

3. Для запуска проекта:

```bash
make run
```

4. API Docs страница (Swagger): [https://cms-lab.gubanov.site/api/docs](https://cms-lab.gubanov.site/api/docs)

![Screenshot](https://user-images.githubusercontent.com/11155743/112715187-07dab100-8ef0-11eb-97ea-68d34f2178f6.png)

## 🗄 Структура шаблона

### ./app

**Папка с бизнес-логикой**. Этот каталог не зависит от _того, какой драйвер базы данных вы используете_ или _какое
решение для кэширования вы выбрали_, или любых других сторонних вещей.

- `./app/controllers` папка для функциональных контроллеров (используются в маршрутах)
- `./app/models` папка для описания бизнес-моделей и методов вашего проекта
- `./app/queries` папка для описания запросов к моделям вашего проекта

### ./docs

**Папка с документацией API**. Этот каталог содержит конфигурационные файлы для автоматически генерируемой документации
API с помощью Swagger.

### ./pkg

**Папка с функциональностью, специфичной для проекта**. Этот каталог содержит весь код, специфичный для вашего
бизнес-кейса, такой как _конфигурации_, _middleware_, _маршруты_ или _утилиты_.

- `./pkg/configs` папка для функций конфигурации
- `./pkg/middleware` папка для добавления middleware (встроенного в Fiber и вашего)
- `./pkg/repository` папка для описания `const` вашего проекта
- `./pkg/routes` папка для описания маршрутов вашего проекта
- `./pkg/utils` папка с утилитарными функциями (запуск сервера, проверка ошибок и т.д.)

### ./platform

**Папка с логикой уровня платформы**. Этот каталог содержит всю логику уровня платформы, которая будет использоваться
для построения фактического проекта, такая как _настройка базы данных_ или _экземпляра кэш-сервера_ и _хранение
миграций_.

- `./platform/cache` папка с функциями настройки in-memory кэша (по умолчанию, Redis)
- `./platform/database` папка с функциями настройки базы данных (по умолчанию, PostgreSQL)
- `./platform/migrations` папка с файлами миграций (используется с
  инструментом [golang-migrate/migrate](https://github.com/golang-migrate/migrate))

## ⚙️ Конфигурация

Смотри пример в файле [.env.test](.env.test)

## ⚙️ JWT токены

Для корректной работы JWT токенов нужно прописать обязательно сгенерированные сертификаты:

```shell
#!/bin/bash

# Генерация JWT ключей и конфига
echo "Generating JWT keys and config..."

# Создаем директорию для сертификатов
mkdir -p ./certs

# Генерируем ключи для access токена
openssl genrsa -out ./certs/access_private.pem 4096
openssl rsa -in ./certs/access_private.pem -pubout -out ./certs/access_public.pem

# Генерируем ключи для refresh токена
openssl genrsa -out ./certs/refresh_private.pem 4096
openssl rsa -in ./certs/refresh_private.pem -pubout -out ./certs/refresh_public.pem

# Читаем ключи и преобразуем в одну строку с \n
access_private=$(awk 'NF {sub(/\r/, ""); printf "%s\\n", $0}' ./certs/access_private.pem)
access_public=$(awk 'NF {sub(/\r/, ""); printf "%s\\n", $0}' ./certs/access_public.pem)
refresh_private=$(awk 'NF {sub(/\r/, ""); printf "%s\\n", $0}' ./certs/refresh_private.pem)
refresh_public=$(awk 'NF {sub(/\r/, ""); printf "%s\\n", $0}' ./certs/refresh_public.pem)

# Создаем конфигурационный файл
cat > ./certs/jwt_config.env <<EOF
# JWT settings:
JWT_SECRET_KEY_PRIVATE="$access_private"
JWT_SECRET_KEY_PUBLIC="$access_public"
JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT=15
JWT_REFRESH_KEY_PRIVATE="$refresh_private"
JWT_REFRESH_KEY_PUBLIC="$refresh_public"
JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT=12
EOF

# Устанавливаем права
chmod -R 755 ./certs

echo "JWT keys and config generated successfully in ./certs/jwt_config.env"
```

Далее, перенести в env переменные сгенерированный файл [./certs/jwt_config.env](./certs/jwt_config.env) и изменить
настройки если необходимо:

```
JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT
JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT
```

Смотри пример в файле [.env.test](.env.test)

Для аддонов используй

```shell
# Addons Configuration
ADDONS_PATH_CONFIG=addon-simple.json
```

```shell
# Подключение к серверам proxmox
PROXMOX_PATH_CONFIG=proxmox-simple.json
```
## ⚙️ Тесты & Линтер

```bash
make test
make lint
pre-commit run --all-files
```

```bash
make test
make lint
pre-commit run --all-files
```

## ⚠️ Лицензия

Apache 2.0 &copy; [Vic Shóstak](https://shostak.dev/) & [True web artisans](https://1wa.co/).
