-- name: CreateUser :one
INSERT INTO app.users (
    name,
    email,
    password_hash
) VALUES ($1, $2, $3)
RETURNING id;


-- name: CheckUserExistsByEmail :one
SELECT EXISTS(
    SELECT 1
    FROM app.users
    WHERE email = $1
);

-- name: GetUserByEmail :one
SELECT id, created_at, name, email, password_hash, activated
FROM app.users
WHERE email = $1;

-- name: ActivateUser :exec
UPDATE app.users
SET activated = true
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE app.users
SET activated = false
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM app.users
WHERE id = $1;
