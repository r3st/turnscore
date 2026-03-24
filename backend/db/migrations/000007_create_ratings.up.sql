CREATE TABLE ratings (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id        UUID        NOT NULL REFERENCES tables(id)  ON DELETE CASCADE,
    rater_id        UUID        NOT NULL REFERENCES raters(id)  ON DELETE CASCADE,
    criteria_scores JSONB       NOT NULL,
    played_zone     VARCHAR(20) NOT NULL DEFAULT 'none'
                                CHECK (played_zone IN ('none', 'zone_a', 'zone_b')),
    comment         TEXT        CHECK (char_length(comment) <= 1000),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (table_id, rater_id)  -- duplicate rating protection at DB level
    -- rater_id is NEVER exposed externally: ratings are anonymous to organizers
);

CREATE INDEX idx_ratings_table_id ON ratings (table_id);
