CREATE TYPE flip_state AS ENUM ('up', 'down', 'paused');

CREATE TABLE IF NOT EXISTS flips
(
    id        int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "to"      flip_state  NOT NULL,
    date      timestamptz NOT NULL,
    check_id  uuid        NOT NULL,
    processed bool        NOT NULL DEFAULT false,
    FOREIGN KEY (check_id) REFERENCES checks (id) ON DELETE CASCADE ON UPDATE CASCADE
);