-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages ADD COLUMN embedded BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages DROP COLUMN embedded;
-- +goose StatementEnd
