CREATE TYPE ping_type AS ENUM ('start', 'success', 'fail');

CREATE TABLE IF NOT EXISTS pings
(
    id           int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type ping_type NOT NULL,
    date    timestamptz NOT NULL,
    source text NOT NULL,
    user_agent text NOT NULL,
    method varchar(4) NOT NULL,
    duration int,
    body varchar(10000),
    check_id uuid NOT NULL,
    FOREIGN KEY (check_id) REFERENCES checks (id) ON DELETE CASCADE ON UPDATE CASCADE
);