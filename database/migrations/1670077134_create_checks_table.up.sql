CREATE TYPE check_status AS ENUM ('new', 'started', 'up', 'down', 'paused');

CREATE TABLE IF NOT EXISTS checks
(
    id           int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name         varchar(255) NOT NULL,
    description  varchar(2000),
    interval     integer      NOT NULL,
    grace        integer      NOT NULL,
    last_ping    timestamptz,
    next_ping    timestamptz,
    last_started timestamptz,
    status       check_status NOT NULL,
    used_id      integer      NOT NULL,
    FOREIGN KEY (used_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
);