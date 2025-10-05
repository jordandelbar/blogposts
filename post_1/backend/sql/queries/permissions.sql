-- name: GetPermissions :many
SELECT p.code
FROM auth.permissions AS p
INNER JOIN auth.users_permissions AS up ON up.permission_id = p.id
INNER JOIN app.users AS u ON up.user_id = u.id
WHERE u.id = $1;

-- name: AddPermissionForUser :exec
INSERT INTO auth.users_permissions (user_id, permission_id)
SELECT $1, p.id
FROM auth.permissions AS p
WHERE p.code = $2;
