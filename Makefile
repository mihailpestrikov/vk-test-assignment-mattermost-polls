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