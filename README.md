![actions status](https://github.com/We-ll-think-about-it-later/identity-service/actions/workflows/ci.yml/badge.svg)

## Запуск

### 1. Создать файл с настройками

```bash
mv env.template .env
```

### 2. Запустить сервис

```bash
go run cmd/app/main.go
```
или, используя `make` и `docker-compose`, запустите сервис вместе с MongoDB:
```bash
make compose-up
```

## Тесты

```bash
go test -v -cover -race ./internal/...
```

или через `make`

```bash
make test
```

## Ручки

### Сервисные ручки

- `GET /healthz` - Kubernetes Liveness probe
- `GET /swagger/index.html` - Документация Swagger

### Аутентификация и авторизация

#### POST /auth/authenticate

**Описание:** Аутентифицирует пользователя. Если пользователь с указанным email существует, отправляется код подтверждения для входа. Если пользователь не существует, создается новый пользователь и отправляется код подтверждения для регистрации.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `email` | string | Адрес электронной почты пользователя | Да |

**Заголовки:**

| Заголовок | Описание |
|---|---|
| `X-Device-Fingerprint` | SHA-256 хэш отпечатка устройства (добавляется на API gateway) |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Пользователь найден | `{ "user_id": "UUID" }` |
| 201 | Пользователь создан | `{ "user_id": "UUID" }` |
| 400 | Неверный запрос | `{ "error": "сообщение об ошибке" }` |

#### POST /auth/token

**Описание:** Получает токены доступа и обновления после ввода кода подтверждения.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `code` | integer | Код подтверждения | Да |

**Заголовки:**

| Заголовок | Описание |
|---|---|
| `X-User-Id` | Идентификатор пользователя (добавляется на API gateway) |
| `X-Device-Fingerprint` | SHA-256 хэш отпечатка устройства (добавляется на API gateway) |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Токены выданы | `{ "access_token": "string", "refresh_token": "string" }` |
| 400 | Неверный запрос | `{ "error": "сообщение об ошибке" }` |
| 401 | `userId` не указан или пользователь не найден | `{ "error": "сообщение об ошибке" }` |
| 403 | Неверный код подтверждения | `{ "error": "сообщение об ошибке" }` |

#### POST /auth/token/refresh

**Описание:** Обновляет токен доступа.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `refresh_token` | string | Токен обновления | Да |

**Заголовки:**

| Заголовок | Описание |
|---|---|
| `X-User-Id` | Идентификатор пользователя (добавляется на API gateway) |
| `X-Device-Fingerprint` | SHA-256 хэш отпечатка устройства (добавляется на API gateway) |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Токен доступа обновлен | `{ "access_token": "string" }` |
| 400 | Неверный запрос | `{ "error": "сообщение об ошибке" }` |
| 401 | `userId` не указан или пользователь не найден | `{ "error": "сообщение об ошибке" }` |
| 403 | Неверный токен обновления | `{ "error": "сообщение об ошибке" }` |

### Управление профилем пользователя

#### POST /users/{user_id}/profile

**Описание:** Создает профиль пользователя.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `username` | string | Никнейм пользователя | Да |
| `firstname` | string | Имя пользователя | Да |
| `lastname` | string | Фамилия пользователя | Нет |

**Заголовки:**

| Заголовок | Описание |
|---|---|
| `X-Device-Fingerprint` | SHA-256 хэш отпечатка устройства (добавляется на API gateway) |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 201 | Профиль пользователя создан |  |
| 400 | Неверный запрос | `{ "error": "сообщение об ошибке" }` |
| 404 | Пользователь не найден | `{ "error": "сообщение об ошибке" }` |
| 409 | Имя пользователя занято | `{ "error": "сообщение об ошибке" }` |

#### PATCH /users/{user_id}/profile

**Описание:** Изменяет профиль пользователя.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `username` | string | Никнейм пользователя | Нет |
| `firstname` | string | Имя пользователя | Нет |
| `lastname` | string | Фамилия пользователя | Нет |

**Заголовки:**

| Заголовок | Описание |
|---|---|
| `X-Device-Fingerprint` | SHA-256 хэш отпечатка устройства (добавляется на API gateway) |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Профиль пользователя обновлен |  |
| 400 | Неверный запрос | `{ "error": "сообщение об ошибке" }` |
| 404 | Профиль пользователя не найден | `{ "error": "сообщение об ошибке" }` |
| 409 | Имя пользователя занято | `{ "error": "сообщение об ошибке" }` |

#### GET /users/{user_id}/profile

**Описание:** Получает профиль пользователя.

**Параметры:**

Нет

**Заголовки:**

| Заголовок | Описание |
|---|---|
| `X-Device-Fingerprint` | SHA-256 хэш отпечатка устройства (добавляется на API gateway) |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Данные пользователя получены | `{ "firstname": "string", "lastname": "string", "email": "string" }` |
| 400 | Неверный запрос | `{ "error": "сообщение об ошибке" }` |
| 404 | Пользователь или профиль пользователя не найдены | `{ "error": "сообщение об ошибке" }` |

