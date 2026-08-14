# Pablo

Публикация контента в Instagram Reels, TikTok, YouTube Shorts и Threads через [zernio](https://zernio.com).

Видео уезжает в три видео-площадки, в Threads пишем текст руками. Публикация — сразу или по расписанию.

## Структура

```
backend/          Go, слоёная структура как в vibik_main
  cmd/service            HTTP API (:8080)
  cmd/publisher_worker   отложенная публикация
  internal/domain        post, target, media, account
  internal/service       бизнес-логика
  internal/storage       postgres (squirrel)
  internal/integration   zernio, s3
  schema/                openapi + сгенерированный chi-server
  migrations/            golang-migrate (.up.sql / .down.sql)
frontend/         React + Vite + TS, в проде раздаётся nginx
```

## Как это работает

1. Видео загружается в S3 **с публичным доступом** — zernio не принимает файл, он скачивает медиа по URL.
2. Пост создаётся с набором таргетов (по одному на площадку), у каждого свой текст и свой статус.
3. Публикация идёт отдельным вызовом zernio на каждую площадку: падение Instagram не отменяет TikTok.
4. Публикация асинхронная — статусы `publishing`/`processing` считаются принятыми, не ошибкой.
5. Отложенные посты забирает `publisher_worker` через `FOR UPDATE SKIP LOCKED`, поэтому пост не уедет дважды.

## Запуск

Заполните `backend/cmd/config.yaml`: ключ zernio, доступы к S3, логин/пароль.
Профиль zernio указывать не нужно — клиент берёт первый существующий, а если профилей нет, создаёт «Pablo» сам.

### Через docker

```bash
docker compose up -d --build
```

Приложение открывается на **http://localhost** — `web` (nginx) раздаёт собранный фронт и проксирует `/v1` на `service`. Наружу торчит только он: API и интерфейс живут на одном origin, поэтому CORS не нужен.

Миграции (первый запуск и после каждой новой):

```bash
docker compose run --rm migrate up
```

`config.yaml` монтируется с хоста, а не копируется в образ: в нём ключи, и правка настроек не должна требовать пересборки. Внутри compose база доступна как `postgres:5432` — именно это и должно стоять в `DB`.

### Локально, без docker

В `config.yaml` поменяйте `DB` на `localhost:5434`, затем:

```bash
cd backend && go run ./cmd/service
```

```bash
cd backend && go run ./cmd/publisher_worker
```

```bash
cd frontend && npm install && npm run dev
```

Vite поднимется на :5173 и проксирует `/v1` на :8080.

## Аккаунты

Аккаунты подключаются через hosted-OAuth zernio: страница «Аккаунты» → «Подключить» → редирект обратно.
Локальная таблица `accounts` — кэш, источник истины остаётся в zernio; «Синхронизировать» обновляет её.

## Генерация API

```bash
cd backend && oapi-codegen -generate chi-server,types -package schema -o schema/schema.gen.go schema/schema.yaml
```
