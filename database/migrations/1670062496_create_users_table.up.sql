CREATE TABLE IF NOT EXISTS users
(
    id       int generated always as identity primary key,
    email    text unique not null,
    password text not null
);