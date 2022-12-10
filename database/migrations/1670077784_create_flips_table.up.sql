CREATE TYPE flip_state AS ENUM ('up', 'down', 'paused');

CREATE TABLE IF NOT EXISTS flips
(
    "to"     flip_state  NOT NULL,
    date     timestamptz NOT NULL,
    check_id int,
    FOREIGN KEY (check_id) REFERENCES checks (id) ON DELETE CASCADE ON UPDATE CASCADE
);