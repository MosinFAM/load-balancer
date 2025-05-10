## балансировщик нагрузки


#### `make build`

Создаёт приложение в bin/redditclone.

#### `make test`

Запускает тесты.

#### `make lint`

Запускает линтер.

## Запуск контейнера

```bash
docker compose -f build/docker-compose.yml up -d --build
```

## Тестирование

```bash
bash test/test.sh
```