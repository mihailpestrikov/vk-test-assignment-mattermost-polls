# Запуск только бота и Tarantool
run:
	docker-compose up -d

# Запуск с Mattermost для разработки
dev:
	docker-compose -f docker-compose.yaml -f docker-compose.dev.yaml up -d

# Остановка всех контейнеров
stop:
	docker-compose -f docker-compose.yaml -f docker-compose.dev.yaml down

# Полная очистка
clean:
	docker-compose -f docker-compose.yaml -f docker-compose.dev.yaml down -v

test-cover:
	go test ./internal/api ./internal/model ./internal/service ./pkg/mattermost -coverprofile=

# Запуск линтера для всех пакетов кроме repository
lint:
	golangci-lint run ./internal/api/... ./internal/model/... ./internal/service/... ./pkg/...

