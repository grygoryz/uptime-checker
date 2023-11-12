CREATE TYPE channel_kind AS ENUM ('email', 'webhook');

CREATE TABLE IF NOT EXISTS channels
(
    id          int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind        channel_kind NOT NULL,
    email       text,
    webhook_url_up text,
    webhook_url_down text,
    user_id     int          NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
);