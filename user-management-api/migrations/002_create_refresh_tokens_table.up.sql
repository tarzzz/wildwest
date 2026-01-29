-- Create refresh_tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL UNIQUE,
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked         BOOLEAN NOT NULL DEFAULT false,
    revoked_at      TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT check_revoked_at CHECK (
        (revoked = true AND revoked_at IS NOT NULL) OR
        (revoked = false AND revoked_at IS NULL)
    )
);

-- Add comment to table
COMMENT ON TABLE refresh_tokens IS 'Stores refresh tokens for JWT authentication, allowing token renewal without re-login';
