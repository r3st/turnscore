CREATE TABLE photos (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id      UUID        NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    url           VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500),
    category      VARCHAR(20) NOT NULL DEFAULT 'general'
                              CHECK (category IN ('general', 'zone_a', 'zone_b')),
    uploaded_by   UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_photos_table_id ON photos (table_id);
