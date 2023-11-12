CREATE TABLE IF NOT EXISTS checks_channels
(
    check_id   uuid NOT NULL,
    channel_id int NOT NULL,
    PRIMARY KEY (check_id, channel_id),
    FOREIGN KEY (check_id) REFERENCES checks (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES channels (id) ON DELETE CASCADE ON UPDATE CASCADE
);