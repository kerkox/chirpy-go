-- +goose Up
CREATE TABLE refresh_tokens (
   token VARCHAR(255) PRIMARY KEY,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   expires_at TIMESTAMPTZ NOT NULL,
   revoked_at TIMESTAMPTZ,
   user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE refresh_tokens;