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
или через make и docker compose сразу вместе с mongo
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

## Endpoints

Этот API предоставляет следующие точки доступа:


### POST /auth/signup

**Описание:** Регистрация нового пользователя. На email отправляется код подтверждения, который необходим для получения токенов через `/auth/get_tokens`.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `firstname` | string | Имя пользователя | Да |
| `lastname` | string | Фамилия пользователя | Нет |
| `email` | string | Адрес электронной почты пользователя | Да |
| `device_fingerprint` | string | Отпечаток устройства (SHA-256 хэш) | Да |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Пользователь успешно зарегистрирован | `{ "user_id": "UUID" }` |
| 400 | Неверный запрос | `{ "error": "сообщение об ошибке" }` |


### POST /auth/login

**Описание:** Вход пользователя. На email отправляется код подтверждения, который необходим для получения токенов через `/auth/get_tokens`.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `email` | string | Адрес электронной почты пользователя | Да |
| `device_fingerprint` | string | Отпечаток устройства (SHA-256 хэш) | Да |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Пользователь найден | `{ "user_id": "UUID" }` |
| 400 | Неверный запрос | `{ "error": "сообщение об ошибке" }` |


### POST /auth/get_tokens

**Описание:** Получение access и refresh токенов после ввода кода подтверждения.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `user_id` | string (UUID) | Идентификатор пользователя | Да |
| `code` | integer | Код подтверждения | Да |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Пользователь подтвержден | `{ "access_token": "string", "refresh_token": "string" }` |


### POST /auth/refresh

**Описание:** Обновление access токена.

**Параметры:**

| Параметр | Тип | Описание | Обязательный |
|---|---|---|---|
| `user_id` | string (UUID) | Идентификатор пользователя | Да |
| `refresh_token` | string | Refresh токен | Да |

**Ответы:**

| Код состояния | Описание | Тело ответа |
|---|---|---|
| 200 | Access токен обновлен | `{ "access_token": "string" }` |
