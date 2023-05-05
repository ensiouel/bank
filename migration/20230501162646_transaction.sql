-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transaction
(
    id         UUID PRIMARY KEY     DEFAULT GEN_RANDOM_UUID(),
    payee_id   UUID REFERENCES balance (user_id) ON DELETE CASCADE,
    payer_id   UUID REFERENCES balance (user_id) ON DELETE CASCADE,
    type       TEXT        NOT NULL,
    amount     BIGINT      NOT NULL,
    comment    TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transaction;
-- +goose StatementEnd
