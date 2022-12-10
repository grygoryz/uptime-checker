CREATE TABLE IF NOT EXISTS users
(
    id       int generated always as identity primary key,
    email    text not null,
    password text not null
);