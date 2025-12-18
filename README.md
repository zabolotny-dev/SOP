# Hosting Service

Сервис для управления тарифными планами и серверами.

## Общая информация

Проект написан на Go в стиле, близком к рекомендациям Ardan Labs:  
код организован по бизнес-доменам (вертикальные срезы), бизнес-логика находится в пакетах `internal/plan` и `internal/server` и не зависит от транспортных слоёв и инфраструктуры.

Транспортные слои (REST API на Chi + oapi-codegen, GraphQL на gqlgen, обработка событий RabbitMQ) подключаются в `main.go` через dependency injection.

Асинхронное выделение IP-адреса реализовано через отдельный сервис `hosting-provisioning-service`, взаимодействие — по RabbitMQ с использованием контрактных структур событий.

## Технологии

- Go 1.24
- PostgreSQL (pgx/v5)
- RabbitMQ
- REST: Chi + oapi-codegen
- GraphQL: gqlgen
- Конфигурация: ardanlabs/conf
- Миграции: goose

## Структура

- `hosting-service/cmd/server` — точка входа и транспортные слои
- `hosting-service/cmd/migrator` — CLI для миграций
- `hosting-service/internal/plan` — бизнес-логика тарифных планов
- `hosting-service/internal/server` — бизнес-логика серверов
- `hosting-service/internal/platform` — общая инфраструктура (БД, middleware)
- `hosting-contracts` — спецификации REST и GraphQL
- `hosting-events-contract` — контракты событий RabbitMQ
- `hosting-kit` — общая обёртка над RabbitMQ
- `hosting-provisioning-service` — отдельный сервис provisioning'а
