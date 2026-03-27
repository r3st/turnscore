CREATE TABLE event_ratings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id   UUID        NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    rater_id        UUID        NOT NULL REFERENCES raters(id)      ON DELETE CASCADE,
    criteria_scores JSONB       NOT NULL DEFAULT '{}',
    comment         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (tournament_id, rater_id)
);
