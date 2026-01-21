-- +goose Up
-- +goose StatementBegin
ALTER TABLE servers ADD COLUMN owner_id UUID NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE servers DROP COLUMN pool_id;
-- +goose StatementEnd
