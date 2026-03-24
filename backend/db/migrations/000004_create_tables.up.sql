CREATE TABLE tables (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID         NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    number        INT          NOT NULL CHECK (number > 0),
    name          VARCHAR(200),
    description   TEXT,
    qr_code_path  VARCHAR(500),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (tournament_id, number)
);

CREATE INDEX idx_tables_tournament_id ON tables (tournament_id);
