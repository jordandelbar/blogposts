CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.permissions (
    id serial PRIMARY KEY,
    code text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS auth.users_permissions (
    user_id bigint NOT NULL REFERENCES app.users ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES auth.permissions ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

INSERT INTO auth.permissions (code)
VALUES
    ('articles:read'),
    ('articles:write')
ON CONFLICT (code) DO NOTHING;
