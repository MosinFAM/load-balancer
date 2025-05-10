## Балансировщик нагрузки

Реализованный функционал:

#### `Балансировка нагрузки`: Round-Robin, health-check, retry, конфигурация через JSON, логирование.

#### `Reverse Proxy`: потокобезопасная маршрутизация с httputil.ReverseProxy.

#### `Rate Limiting`: Token Bucket с granular-лимитами (по IP/API-ключу), хранение в PostgreSQL, 429-ответы.

#### `Инфраструктура`: Docker, docker-compose, юнит-тесты.



#### `make build`

Собрать бинарник в bin/redditclone.

#### `make test`

Запуск тестов.

#### `make lint`

Запуск линтера.

## Запуск контейнера

```bash
docker compose -f build/docker-compose.yml up -d --build
```

## Тестирование производительности

```bash
bash test/test.sh
```