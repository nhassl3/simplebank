# Simplebank

REST API банковского приложения на Go: пользователи, счета (USD/EUR), переводы и роли. JWT/PASETO, PostgreSQL, миграции, sqlc.

---

## Возможности

- **Пользователи** — регистрация, вход, профиль (username, full name, email), смена пароля и имени
- **Счета** — создание, просмотр, список, изменение баланса, удаление (владелец проверяется по токену)
- **Переводы** — переводы между счетами с проверкой валюты и прав доступа
- **Роли** — обычный пользователь и админ (доступ к данным других пользователей по `username`)
- **Безопасность** — Bearer JWT/PASETO, argon2 для паролей, CORS, валидация запросов

---

## Стек

| Категория   | Технологии |
|------------|------------|
| Язык       | Go 1.25+   |
| HTTP       | Gin        |
| БД         | PostgreSQL 18, pgx/v5 |
| Миграции   | golang-migrate |
| Запросы    | sqlc       |
| Токены     | JWT, PASETO |
| Пароли     | argon2id   |
| Конфиг     | YAML + .env, cleanenv, viper |

---

## Требования

- Go 1.25+
- PostgreSQL 18 (локально или Docker)
- [migrate](https://github.com/golang-migrate/migrate) (CLI)
- [sqlc](https://sqlc.dev/) (для генерации кода)

---

## Быстрый старт

### Локально

1. Клонировать репозиторий и перейти в каталог:
   ```bash
   git clone https://github.com/nhassl3/simplebank.git && cd simplebank
   ```

2. Создать `.env` в корне проекта:
   ```env
   DATABASE_PASSWORD=your_password
   CONNECTION_STRING=postgres://root:your_password@localhost:5432/simple_bank?sslmode=disable
   TOKEN_SECRET_KEY_GENERATION=your_secret_key_min_32_chars
   ```

3. Поднять PostgreSQL (например, через Makefile):
   ```bash
   make postgres
   make createdb
   make migrate-up
   ```

4. Запуск приложения:
   ```bash
   make run
   ```
   По умолчанию API доступен на `http://localhost:8080`.

### Docker Compose

```bash
# В корне проекта должны быть prod.env с переменными (DATABASE_PASSWORD, CONNECTION_STRING, TOKEN_SECRET_KEY_GENERATION и т.д.)
docker compose up
```

API — порт `8080`, фронтенд (если собран) — `3000`, PostgreSQL — внутренняя сеть.

---

## Конфигурация

- **Конфиг приложения** — YAML: `config/local.yaml` (или `config/prod.yaml`). Указывается флагом `--config=./config/local.yaml`.
- **Секреты и строка подключения** — из `.env`, путь к файлу: `--env=.env`.

Пример локального конфига:

```yaml
# config/local.yaml
log_type: 1
http:
  host: "localhost"
  port: 8080
tgp:
  access_token_duration: "15m"
```

Переменные из `.env`: `CONNECTION_STRING`, `TOKEN_SECRET_KEY_GENERATION`, при необходимости `DATABASE_PASSWORD` для CLI миграций.

---

## API

Базовый URL: `http://localhost:8080` (или хост/порт из конфига).

### Аутентификация (без токена)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/auth/signup` | Регистрация (username, password, full_name, email) |
| POST | `/api/auth/login`  | Вход (username, password) → токен |

Дальнейшие запросы к защищённым эндпоинтам требуют заголовок:
`Authorization: Bearer <access_token>`.

### API v1 (с токеном)

**Счета** — `/api/v1/accounts`

| Метод | Путь | Описание |
|-------|------|----------|
| POST   | `/`        | Создать счёт (owner, currency). Обычный пользователь — только для себя |
| GET    | `/:id`     | Получить счёт по ID (только свой) |
| GET    | `/`        | Список счетов (query: page, limit) |
| PUT    | `/`        | Обновить баланс (id, balance) |
| PUT    | `/addBalance` | Изменить баланс на величину (id, amount) |
| DELETE | `/:id`     | Удалить счёт (только свой) |

**Переводы** — `/api/v1/transfers`

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/` | Создать перевод (from_account_id, to_account_id, amount, currency). Счёт отправителя должен принадлежать текущему пользователю |

**Пользователи** — `/api/v1/users`

| Метод | Путь | Описание |
|-------|------|----------|
| GET    | `/:username`       | Профиль пользователя. Админ — любой username |
| PUT    | `/update/password` | Смена пароля (username, password) |
| PUT    | `/update/fullname` | Смена имени (username, full_name) |
| DELETE | `/:username`       | Удаление пользователя. Админ — любой username |

Роль админа задаётся в БД (`level_right` и т.п.) и отражается в токене; админ может указывать чужой `username` в запросах.

---

## База данных

- Схема: пользователи (`users`), счета (`accounts`), записи движений (`entries`), переводы (`transfers`). Связь `accounts.owner` → `users.username`.
- Миграции: `internals/db/migration/` (golang-migrate).

Команды:

```bash
make migrate-up      # применить миграции
make migrate-down    # откатить
make migrate-up-once # применить одну миграцию вверх
make migrate-down-once
make opendb          # psql в контейнер (при запущенном postgres через make postgres)
```

Диаграмма схемы: [SVG](.github/source/simplebank_db_diagram.svg) / [PDF](.github/source/simplebank_db_diagram.pdf).

---

## Разработка

| Команда | Описание |
|--------|----------|
| `make run`   | Сборка и запуск с `config/local.yaml` и `.env` |
| `make build` | Сборка бинарника в `build/` |
| `make test`  | Запуск тестов |
| `make sqlc`  | Генерация кода из `internals/db/query/*.sql` |
| `make mock`  | Генерация моков для Store |
| `make clean` | Удаление `build/` |

После изменения SQL-запросов выполните `make sqlc`. Конфигурация sqlc — `sqlc.yaml`.

---

## Структура проекта

```
simplebank/
├── cmd/simplebank/          # Точка входа
├── config/                 # YAML конфиги (local, prod)
├── internals/
│   ├── app/                # Инициализация приложения и БД
│   ├── config/             # Загрузка конфигурации
│   ├── db/
│   │   ├── migration/      # SQL миграции
│   │   ├── query/          # SQL запросы для sqlc
│   │   └── sqlc/           # Сгенерированный код и моки
│   ├── domain/http/        # Обработчики бизнес-логики (handlers)
│   ├── http/
│   │   ├── middleware/     # Auth, проверка пользователя
│   │   └── simplebank/     # Роутер, хендлеры HTTP, сессии/запросы
│   └── lib/                # Логгер, токены, валидатор
├── tests/                  # API-тесты
├── docker-compose.yaml
├── Makefile
├── sqlc.yaml
└── README.md
```

---

## CI

В GitHub Actions при пуше/PR в `main` запускаются тесты (Go 1.25.5, PostgreSQL 18, миграции и `make test`). Конфиг: [.github/workflows/go.yml](.github/workflows/go.yml).

