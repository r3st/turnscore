CREATE TABLE raters (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID        NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    nickname      VARCHAR(100) NOT NULL,
    code          VARCHAR(10)  NOT NULL CHECK (code ~ '^\d{4,6}$'),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tournament_id, nickname),
    UNIQUE (tournament_id, code)
    -- No email, no full name: GDPR compliant
);

CREATE INDEX idx_raters_tournament_id ON raters (tournament_id);
