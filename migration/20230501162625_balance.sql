-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS balance
(
    user_id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    balance BIGINT NOT NULL  DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "user";
-- +goose StatementEnd
