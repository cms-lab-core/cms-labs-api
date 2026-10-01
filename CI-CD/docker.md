# Документация по настройке и запуску окружения

Данная документация описывает структуру проекта, настройку и запуск окружения с использованием Docker Compose. Проект
состоит из 4 основных компонентов: `core-backend`, `clabgate`, `mysqldatabase` и `core-frontend`, которые
взаимодействуют через общую сеть `cms-net`.

---

## Структура проекта

Проект организован в следующей структуре директорий:

```
user@vm:~$ tree
.
├── core
│   └── docker-compose.yml
├── clabgate
│   └── docker-compose.yml
├── db
│   └── docker-compose.yml
├── docs
└── front
    └── docker-compose.yml
```

- **core** — содержит конфигурацию для запуска backend-сервиса.
- **clabgate** — содержит конфигурацию для запуска clabgate proxy k8s-сервиса.
- **db** — содержит конфигурацию для запуска базы данных MySQL.
- **front** — содержит конфигурацию для запуска frontend-сервиса.
- **docs** — директория для документации (опционально).

---

## Примеры `docker-compose.yml`

### 1. Backend (`core/docker-compose.yml`)

[⚙️ JWT токены](../backend/README.md#-jwt-токены)

```yaml
version: '3.9'

services:
  core-backend:
    platform: linux/x86_64
    container_name: core-backend
    image: ghcr.io/cms-lab-core/cms-labs-api/backend:${VERSION}
    networks:
      - cms-net
    environment:
      STAGE_STATUS: "prod"
      DEBUG: "false"
      # Server settings:
      SERVER_HOST: "0.0.0.0"
      SERVER_PORT: "5000"
      SERVER_READ_TIMEOUT: "60"

      # JWT settings:
      JWT_SECRET_KEY_PRIVATE: "-----BEGIN PRIVATE KEY-----\n"
      JWT_SECRET_KEY_PUBLIC: "-----BEGIN PUBLIC KEY-----\n"
      JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT: "15"
      JWT_REFRESH_KEY_PRIVATE: "-----BEGIN PRIVATE KEY-----\n"
      JWT_REFRESH_KEY_PUBLIC: "-----BEGIN PUBLIC KEY-----\n"
      JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT: "12"

      # Database settings:
      DB_TYPE: "mysql"
      DB_HOST: "mysqldatabase"
      DB_PORT: "3306"
      DB_USER: "user"
      DB_PASSWORD: "passwd"
      DB_NAME: "db"
      DB_SSL_MODE: "disable"
      DB_MAX_CONNECTIONS: "100"
      DB_MAX_IDLE_CONNECTIONS: "10"
      DB_MAX_LIFETIME_CONNECTIONS: "30"
      DB_TABLE_PREFIX: "backend_"

networks:
  cms-net:
    external: true
```

### 2. Clabgate (`clabgate/docker-compose.yml`)

```yaml
version: '3.9'

services:
  clabgate:
    platform: linux/x86_64
    container_name: clabgate
    image: ghcr.io/cms-lab-core/cms-labs-api/clabgate:${VERSION}
    networks:
      - cms-net
    environment:
      STAGE_STATUS: "prod"
      DEBUG: "false"
      # Server settings:
      SERVER_HOST: "0.0.0.0"
      SERVER_PORT: "5001"
      SERVER_READ_TIMEOUT: "60"

      # Task catalog (GitLab or GitHub)
      CMS_TASK_URL: "https://github.com/example/lab-tasks"
      CMS_TASK_BRANCH: "main"
      TASK_REPOSITORY_TOKEN: ""

networks:
  cms-net:
    external: true
```

### 2. База данных (`db/docker-compose.yml`)

```yaml
version: '3.9'

services:
  mysqldatabase:
    container_name: mysqldatabase
    image: mysql:9.1
    restart: always
    environment:
      MYSQL_DATABASE: 'db'
      MYSQL_USER: 'user'
      MYSQL_PASSWORD: 'passwd'
      MYSQL_ROOT_PASSWORD: 'root_passwd'
    networks:
      - cms-net
    ports:
      - '33306:3306'
    volumes:
      - ./local_db:/var/lib/mysql
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

networks:
  cms-net:
    external: true
```

### 3. Frontend (`front/docker-compose.yml`)

```yaml
version: '3.9'

services:
  core-frontend:
    platform: linux/x86_64
    container_name: core-frontend
    image: ghcr.io/cms-lab-core/cms-labs-api/frontend:${VERSION}
    networks:
      - cms-net
    ports:
      - "5200:80"
    environment:
      CMS_CORE_SERVICE: "http://core-backend:5000"
      CLABGATE_SERVICE: "http://clabgate:5000"
      SYSTEM_DNS: "127.0.0.11"
      SSL_CONF: ""
  #      SSL_CONF: "include ssl.conf;"
  #    volumes:
  #      - ./certs/:/etc/certs

networks:
  cms-net:
    external: true
```

---

## Запуск проекта

### 1. Создание сети `cms-net`

Перед запуском контейнеров необходимо создать Docker-сеть `cms-net`:

```bash
docker network create cms-net
```

### 2. Запуск базы данных

Перейдите в директорию `db` и запустите контейнер с базой данных:

```bash
cd db
docker-compose up -d
```

### 3. Запуск backend-сервиса

Перейдите в директорию `core` и запустите контейнер с backend-сервисом:

```bash
cd ../core
docker-compose up -d
```

### 4. Запуск frontend-сервиса

Перейдите в директорию `front` и запустите контейнер с frontend-сервисом:

```bash
cd ../front
docker-compose up -d
```

---

## Взаимодействие между сервисами

- **Frontend** взаимодействует с **Backend** через адрес `http://core-backend:5000`.
- **Backend** взаимодействует с **MySQL** через адрес `mysqldatabase:3306`.

---

## Остановка и удаление контейнеров

Для остановки и удаления контейнеров выполните следующие команды в каждой из директорий (`core`, `db`, `front`):

```bash
docker-compose down
```

---

## Дополнительные настройки

### SSL для Frontend

Для настройки SSL в frontend-сервисе:

1. Создайте директорию `certs` в папке `front`.
2. Поместите сертификаты в эту директорию.
3. Раскомментируйте строки в `docker-compose.yml` для подключения SSL:

```yaml
volumes:
  - ./certs/:/etc/certs
environment:
  SSL_CONF: "include ssl.conf;"
```

### Изменение версий образов

Для изменения версии образов (backend или frontend) задайте переменную окружения `VERSION` перед запуском контейнеров:

```bash
export VERSION=1.0.0
docker-compose up -d
```

---

## Логирование

Для просмотра логов контейнеров используйте команду:

```bash
docker logs <container_name>
```

Например:

```bash
docker logs core-backend
```

---

## Заключение

Данная документация описывает структуру проекта, настройку и запуск окружения с использованием Docker Compose. Для более
детальной настройки (например, SSL, мониторинг, логирование) обратитесь к документации Docker и используемым образам.
