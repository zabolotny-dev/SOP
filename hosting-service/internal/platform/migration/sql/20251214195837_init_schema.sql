-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    cpu_cores INT NOT NULL,
    ram_mb INT NOT NULL,
    disk_gb INT NOT NULL
);

CREATE TABLE IF NOT EXISTS servers (
    id UUID PRIMARY KEY,
    plan_id UUID NOT NULL REFERENCES plans(id),
    name TEXT NOT NULL,
    ipv4_address TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS servers;
DROP TABLE IF EXISTS plans;
-- +goose StatementEnd