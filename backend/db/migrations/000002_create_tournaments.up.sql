CREATE TABLE tournaments (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            VARCHAR(100) UNIQUE NOT NULL,
    name            VARCHAR(200) NOT NULL,
    type            VARCHAR(20)  NOT NULL CHECK (type IN ('fantasy', 'scifi')),
    description     TEXT,
    links           JSONB        NOT NULL DEFAULT '[]',
    location        VARCHAR(200),
    event_date      DATE,
    organizer_id    UUID         NOT NULL REFERENCES users(id),
    table_count     INT          NOT NULL CHECK (table_count > 0),
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft'
                                 CHECK (status IN ('draft', 'active', 'voting', 'archived')),
    voting_start    TIMESTAMPTZ,
    voting_end      TIMESTAMPTZ,
    active_criteria JSONB        NOT NULL DEFAULT '["balance","aesthetics","terrain_density","labeling","overall","zone_a","zone_b"]',
    result_config   JSONB        NOT NULL DEFAULT '{"show_comments":false,"visible_comment_criteria":[]}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tournaments_slug         ON tournaments (slug);
CREATE INDEX idx_tournaments_organizer_id ON tournaments (organizer_id);
CREATE INDEX idx_tournaments_status       ON tournaments (status);
CREATE INDEX idx_tournaments_event_date   ON tournaments (event_date);
