-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN hashed_password TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SELECT 'down SQL query';
-- +goose StatementEnd
