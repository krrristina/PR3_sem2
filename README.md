# Практическая работа №3 (семестр 2)

## Выполнила: Сорокина К.С., ЭФМО-01-25

## Тема: Логирование с помощью zap. Ведение структурированных логов

### Цель:

Научиться внедрять структурированные логи в сервис и применять единый стандарт логирования для диагностики и эксплуатации.

## Технологии

- **Go** — язык реализации
- **zap** — библиотека структурированного логирования
- **middleware** — промежуточные обработчики HTTP-запросов
- **X-Request-ID** — корреляция запросов между сервисами
- **gRPC metadata** — передача request-id в межсервисных вызовах

## Структура проекта

```
PR3_sem2/
├── go.mod
├── go.sum
├── proto/
│   ├── auth.proto
│   ├── auth.pb.go
│   └── auth_grpc.pb.go
├── shared/
│   ├── logger/
│   │   └── logger.go          — инициализация логгера
│   └── middleware/
│       ├── requestid.go       — получение/генерация X-Request-ID
│       └── accesslog.go       — логирование HTTP-запросов
└── services/
    ├── auth/
    │   ├── cmd/auth/main.go
    │   └── internal/grpc/server.go
    └── tasks/
        ├── cmd/tasks/main.go
        └── internal/handler.go
```

## Выбор логгера

Для реализации выбран **zap** (go.uber.org/zap) по следующим причинам:

- Пишет логи в формате JSON по умолчанию — удобно для фильтрации и поиска
- Высокая производительность — не использует reflection
- Поддерживает структурированные поля (`zap.String`, `zap.Int`, `zap.Error`)
- Легко добавить общее поле `service` ко всем логам сразу

## Стандарт полей логов

Во всех сервисах используется единый набор полей:

| Поле | Описание | Пример |
|---|---|---|
| `level` | Уровень лога | `info`, `warn`, `error` |
| `ts` | Время события | `2026-05-03T22:16:32.257+0300` |
| `service` | Имя сервиса | `auth`, `tasks` |
| `request_id` | Идентификатор запроса | `pz3-004` |
| `method` | HTTP метод | `GET`, `POST` |
| `path` | Путь запроса | `/tasks` |
| `status` | Код ответа | `200`, `401`, `503` |
| `duration_ms` | Длительность обработки | `3` |
| `error` | Текст ошибки (без секретов) | `invalid token` |
| `component` | Компонент/слой | `auth_client`, `handler` |

## Команды для запуска сервисов

### Auth (gRPC-сервер)

```bash
cd PR3_sem2
export AUTH_GRPC_PORT=50051
go run ./services/auth/cmd/auth
```

### Tasks (HTTP + gRPC-клиент)

```bash
cd PR3_sem2
export TASKS_PORT=8082
export AUTH_GRPC_ADDR=localhost:50051
go run ./services/tasks/cmd/tasks
```

## Проверка через Postman

### 1. Успешный запрос

**Запрос:**

- Method: `GET`
- URL: `http://localhost:8082/tasks`
- Headers: `Authorization: Bearer my-test-token`

![](https://github.com/krrristina/PR3_sem2/blob/main/screenshots/GET%20tasks.png)
![](https://github.com/krrristina/PR3_sem2/blob/main/screenshots/token%20is%20valid.png)
---

### 2. Запрос с невалидным токеном

**Запрос:**

- Method: `GET`
- URL: `http://localhost:8082/tasks`
- Headers: `Authorization: Bearer invalid-token`

В логах tasks виден `warn` с компонентом `auth_client`, клиент получает `401 Unauthorized`.

![](https://github.com/krrristina/PR3_sem2/blob/main/screenshots/unauthorized.png)
![](https://github.com/krrristina/PR3_sem2/blob/main/screenshots/unauthorized%202.png)
---

### 3. Запрос без токена

**Запрос:**

- Method: `GET`
- URL: `http://localhost:8082/tasks`

В логах tasks виден `warn` с компонентом `handler`, клиент получает `401 missing token`.

![](https://github.com/krrristina/PR3_sem2/blob/main/screenshots/запрос%20без%20токена.png)
---

### 4. Межсервисный вызов (корреляция через request-id)

**Запрос:**

- Method: `GET`
- URL: `http://localhost:8082/tasks`
- Headers: `Authorization: Bearer my-test-token`, `X-Request-ID: pz3-004`

Один и тот же `request_id: pz3-004` виден в логах **обоих** сервисов одновременно — это подтверждает корреляцию запросов между tasks и auth.

![](https://github.com/krrristina/PR3_sem2/blob/main/screenshots/одинаковый%20request%20id%20tasks.png)
![](https://github.com/krrristina/PR3_sem2/blob/main/screenshots/одинаковый%20request%20id%20auth.png)
---

## Примеры лог-событий

### Успешный запрос (tasks)
```json
{"level":"info","ts":"2026-05-03T22:16:32.257+0300","msg":"request completed","service":"tasks","request_id":"pz3-004","method":"GET","path":"/tasks","status":200,"duration_ms":3}
```

### Запрос с невалидным токеном (auth)
```json
{"level":"warn","ts":"2026-05-03T22:16:32.257+0300","msg":"invalid token","service":"auth","request_id":"pz3-002"}
```

### Межсервисный вызов (tasks → auth, одинаковый request_id)
```json
{"level":"info","ts":"2026-05-03T22:16:32.254+0300","msg":"calling grpc verify","service":"tasks","request_id":"pz3-004","component":"auth_client"}
{"level":"info","ts":"2026-05-03T22:16:32.255+0300","msg":"verify called","service":"auth","request_id":"pz3-004","has_token":true}
```

---

## Контрольные вопросы

### Вопрос 1. Почему структурированные логи удобнее строковых?

Структурированные логи содержат набор полей (ключ-значение), а не просто текст. Это позволяет фильтровать и искать по конкретному полю в системах мониторинга (ELK, Loki, Datadog) — например, найти все запросы с `status=500` или все события по `request_id=pz3-004`. Строковые логи для этого пришлось бы парсить через регулярные выражения, что медленно и ненадёжно.

---

### Вопрос 2. Что такое request-id и как он помогает при диагностике?

Request-id — это уникальный идентификатор запроса, который присваивается при входе в первый сервис и передаётся во все последующие сервисы. Благодаря ему можно найти все лог-события одного запроса сразу в нескольких сервисах. Без request-id при ошибке в цепочке из трёх сервисов невозможно понять какие именно события в каждом сервисе относятся к одному и тому же запросу.

---

### Вопрос 3. Какие поля вы считаете обязательными для access log?

Обязательными считаю: `request_id`, `method`, `path`, `status`, `duration_ms`. Эти поля позволяют понять что за запрос пришёл, как он завершился и сколько времени занял — минимально необходимый набор для диагностики.

---

### Вопрос 4. Почему нельзя писать токены и пароли в логи?

Логи хранятся в незашифрованном виде и доступны широкому кругу людей — разработчикам, DevOps-инженерам, системам мониторинга. Если токен попадёт в лог, любой кто имеет доступ к логам сможет использовать его для авторизации от имени пользователя. Вместо значения токена логируется только факт его наличия: `has_token: true`.

---

### Вопрос 5. Что логировать в ERROR, а что в INFO/WARN?

- **ERROR** — непредвиденные сбои, которые требуют внимания: недоступна база данных, паника, таймаут при вызове внешнего сервиса.
- **WARN** — нежелательные, но ожидаемые ситуации: невалидный токен, отсутствие авторизации, устаревший endpoint.
- **INFO** — нормальный жизненный цикл: запрос получен, запрос завершён, сервис запустился.
