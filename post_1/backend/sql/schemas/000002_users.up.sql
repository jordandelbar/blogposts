CREATE EXTENSION IF NOT EXISTS citext;

CREATE SCHEMA IF NOT EXISTS app;

CREATE TABLE app.users (
    id serial PRIMARY KEY,
    created_at timestamp(0) with time zone DEFAULT now(),
    name text NOT NULL,
    email citext NOT NULL UNIQUE,
    password_hash bytea NOT NULL,
    activated bool NOT NULL DEFAULT false
);
