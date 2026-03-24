CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    google_sub         VARCHAR(255) UNIQUE NOT NULL,
    email              VARCHAR(255) UNIQUE NOT NULL,
    name               VARCHAR(200) NOT NULL,
    avatar_url         VARCHAR(500),
    helper_invite_code VARCHAR(20)  UNIQUE NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- NO password_hash: login via Google OAuth only
    -- NO global role: role is per-tournament in tournament_members
);

CREATE INDEX idx_users_google_sub         ON users (google_sub);
CREATE INDEX idx_users_helper_invite_code ON users (helper_invite_code);
