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
  migrations/            goose-совместимые миграции
frontend/         React + Vite + TS
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

Миграции:

```bash
goose -dir backend/migrations postgres "host=localhost port=5434 user=pablo password=pablo dbname=pablo sslmode=disable" up
```

Бэкенд:

```bash
cd backend && go run ./cmd/service
```

Воркер:

```bash
cd backend && go run ./cmd/publisher_worker
```

Фронт:

```bash
cd frontend && npm install && npm run dev
```

Всё вместе:

```bash
docker compose up --build
```

## Аккаунты

Аккаунты подключаются через hosted-OAuth zernio: страница «Аккаунты» → «Подключить» → редирект обратно.
Локальная таблица `accounts` — кэш, источник истины остаётся в zernio; «Синхронизировать» обновляет её.

## Генерация API

```bash
cd backend && oapi-codegen -generate chi-server,types -package schema -o schema/schema.gen.go schema/schema.yaml
```
