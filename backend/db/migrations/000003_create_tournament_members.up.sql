CREATE TABLE tournament_members (
    tournament_id UUID        NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id       UUID        NOT NULL REFERENCES users(id)       ON DELETE CASCADE,
    role          VARCHAR(20) NOT NULL CHECK (role IN ('organizer', 'helper')),
    joined_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tournament_id, user_id)
);

CREATE INDEX idx_tournament_members_user_id ON tournament_members (user_id);
