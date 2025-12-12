# Hosting Service

Сервис для управления тарифными планами и серверами.
Проект реализован с использованием архитектурного подхода **Package Oriented Design** и принципов **Clean Architecture**.

## Архитектура

Приложение построено как модульный монолит. Кодовая база организована вокруг бизнес-доменов, а не технических слоев.

**Ключевые архитектурные решения:**

*   **Package Oriented Design:** Логика инкапсулирована в доменных пакетах (`internal/server`, `internal/plan`). Зависимости направлены внутрь, бизнес-логика не зависит от транспорта.
*   **Hexagonal Architecture:** Транспортный слой (REST, GraphQL, RabbitMQ) отделен от бизнес-логики. Это позволяет менять протоколы взаимодействия без изменения ядра приложения.
*   **Dependency Injection:** Явное внедрение зависимостей в `main.go`. Глобальное состояние отсутствует.

## Технологический стек

*   **Язык:** Go 1.24.4
*   **База данных:** PostgreSQL (драйвер `pgx/v5`)
*   **Брокер сообщений:** RabbitMQ
*   **API:**
    *   **REST:** Chi Router + OAPI Codegen (OpenAPI v3)
    *   **GraphQL:** gqlgen
*   **Конфигурация:** ardanlabs/conf

## Структура проекта

*   `hosting-service/cmd/server/` — Точка входа, конфигурация и транспортный слой.
*   `hosting-service/internal/plan/` — Домен "Тарифные планы".
*   `hosting-service/internal/server/` — Домен "Серверы".
*   `hosting-service/internal/platform/` — Базовые технические возможности (Database connection, Middleware).
*   `hosting-contracts/` — Спецификации API (OpenAPI, GraphQL schemas).
*   `hosting-events-contract/` — Контракты асинхронного взаимодействия (структуры событий, топология очередей).
*   `hosting-kit/` — Общие инфраструктурные библиотеки (обертка над RabbitMQ).
